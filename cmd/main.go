package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/grassrootseconomics/celo-tracker/internal/api"
	"github.com/grassrootseconomics/celo-tracker/internal/backfiller"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/pool"
	"github.com/grassrootseconomics/celo-tracker/internal/processor"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/internal/queue"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
	"github.com/grassrootseconomics/celo-tracker/internal/syncer"
	"github.com/knadh/koanf/v2"
)

const (
	defaultGracefulShutdownPeriod = time.Second * 30

	// 24 hrs worth of blocks
	defaultMaxQueueSize = 17_280
)

var (
	build = "dev"

	confFlag string

	lo *slog.Logger
	ko *koanf.Koanf
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.Parse()

	lo = initLogger()
	ko = initConfig()

	lo.Info("starting celo tracker", "build", build)
}

/*
Dependency Order
----------------
- Chain
- DB
- Cache
- JetStream Pub
- Worker Pool
- Block Processor
- Queue
- Stats
- Chain Syncer
- Backfiller
- API
*/
func main() {
	var wg sync.WaitGroup
	ctx, stop := notifyShutdown()

	chain, err := chain.NewRPCFetcher(chain.RPCOpts{
		RPCEndpoint:   ko.MustString("chain.rpc_endpoint"),
		ChainID:       ko.MustInt64("chain.chainid"),
		IsArchiveNode: ko.Bool("chain.archive_node"),
	})
	if err != nil {
		lo.Error("could not initialize chain client", "error", err)
		os.Exit(1)
	}

	db, err := db.New(db.DBOpts{
		Logg:   lo,
		DBType: ko.MustString("core.db_type"),
	})
	if err != nil {
		lo.Error("could not initialize blocks db", "error", err)
		os.Exit(1)
	}

	cache, err := cache.New(cache.CacheOpts{
		Chain:      chain,
		Logg:       lo,
		CacheType:  ko.MustString("core.cache_type"),
		Blacklist:  ko.MustStrings("bootstrap.blacklist"),
		Registries: ko.MustStrings("bootstrap.ge_registries"),
		Watchlist:  ko.MustStrings("bootstrap.watchlist"),
	})
	if err != nil {
		lo.Error("could not initialize cache", "error", err)
		os.Exit(1)
	}

	jetStreamPub, err := pub.NewJetStreamPub(pub.JetStreamOpts{
		Endpoint:        ko.MustString("jetstream.endpoint"),
		PersistDuration: time.Duration(ko.MustInt("jetstream.persist_duration_hrs")) * time.Hour,
		DedupDuration:   time.Duration(ko.MustInt("jetstream.dedup_duration_hrs")) * time.Hour,
		Logg:            lo,
	})
	if err != nil {
		lo.Error("could not initialize jetstream pub", "error", err)
		os.Exit(1)
	}

	poolOpts := pool.PoolOpts{
		Logg:        lo,
		WorkerCount: ko.Int("core.pool_size"),
		// Immidiately allow processing of upto 6 hrs of missing blocks
		BlocksBuffer: defaultMaxQueueSize / 4,
	}
	if ko.Int("core.pool_size") <= 0 {
		// TODO: Benchamrk to determine optimum size
		poolOpts.WorkerCount = runtime.NumCPU() * 3
	}
	workerPool := pool.NewPool(poolOpts)

	stats := stats.New(stats.StatsOpts{
		Cache: cache,
		Logg:  lo,
		Pool:  workerPool,
	})

	blockProcessor := processor.NewProcessor(processor.ProcessorOpts{
		Cache: cache,
		DB:    db,
		Chain: chain,
		Pub:   jetStreamPub,
		Logg:  lo,
		Stats: stats,
	})

	queue := queue.New(queue.QueueOpts{
		Logg:      lo,
		Processor: blockProcessor,
		Pool:      workerPool,
	})

	chainSyncer, err := syncer.New(syncer.SyncerOpts{
		DB:                db,
		Chain:             chain,
		Logg:              lo,
		Queue:             queue,
		Stats:             stats,
		StartBlock:        ko.Int64("chain.start_block"),
		WebSocketEndpoint: ko.MustString("chain.ws_endpoint"),
	})
	if err != nil {
		lo.Error("could not initialize chain syncer", "error", err)
		os.Exit(1)
	}

	backfiller := backfiller.New(backfiller.BackfillerOpts{
		MaxQueueSize: defaultMaxQueueSize,
		DB:           db,
		Logg:         lo,
		Queue:        queue,
	})

	apiServer := &http.Server{
		Addr:    ko.MustString("api.address"),
		Handler: api.New(stats),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		queue.Process()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		chainSyncer.Start()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := backfiller.Run(false); err != nil {
			lo.Error("backfiller initial run error", "error", err)
		}
		backfiller.Start()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("metrics and stats server starting", "address", ko.MustString("api.address"))
		if err := apiServer.ListenAndServe(); err != http.ErrServerClosed {
			lo.Error("failed to start API server", "error", err)
			os.Exit(1)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		stats.StartStatsPrinter()
	}()

	<-ctx.Done()
	lo.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		queue.Stop()
		workerPool.StopAndWait()
		stats.Stop()
		chainSyncer.Stop()
		backfiller.Stop()
		jetStreamPub.Close()
		db.Cleanup()
		db.Close()
		apiServer.Shutdown(shutdownCtx)
		lo.Info("graceful shutdown routine complete")
	}()

	go func() {
		wg.Wait()
		stop()
		cancel()
		os.Exit(0)
	}()

	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		stop()
		cancel()
		lo.Error("graceful shutdown period exceeded, forcefully shutting down")
	}
	os.Exit(1)
}

func notifyShutdown() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
}
