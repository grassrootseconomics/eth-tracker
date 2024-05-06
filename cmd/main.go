package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/ef-ds/deque/v2"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/processor"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
	"github.com/grassrootseconomics/celo-tracker/internal/syncer"
	"github.com/grassrootseconomics/celo-tracker/pkg/chain"
	"github.com/knadh/koanf/v2"
)

const defaultGracefulShutdownPeriod = time.Second * 20

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

func main() {
	// mux := http.NewServeMux()
	// statsviz.Register(mux)

	// go func() {
	// 	lo.Info("metrics", "host:port", http.ListenAndServe("localhost:8080", mux))
	// }()
	// go func() {
	// 	lo.Info("profiler", "host:port", http.ListenAndServe("localhost:6060", nil))
	// }()

	var (
		batchQueue  deque.Deque[uint64]
		blocksQueue deque.Deque[types.Block]
		wg          sync.WaitGroup
	)
	ctx, stop := notifyShutdown()

	/*
		Dependency Order
		----------------
		- Stats
		- Chain
		- DB
		- HistoricalSyncer
		- RealtimeSyncer
		- BlockProcessor
	*/
	stats := stats.New(lo)

	chain, err := chain.New(chain.ChainOpts{
		RPCEndpoint: ko.MustString("chain.rpc_endpoint"),
		TestNet:     ko.Bool("chain.testnet"),
		Logg:        lo,
	})
	if err != nil {
		lo.Error("could not initialize chain client", "error", err)
		os.Exit(1)
	}

	db, err := db.New(db.DBOpts{
		Logg: lo,
	})
	if err != nil {
		lo.Error("could not initialize blocks db", "error", err)
		os.Exit(1)
	}

	chainSyncer, err := syncer.New(syncer.SyncerOpts{
		WebSocketEndpoint: ko.MustString("chain.ws_endpoint"),
		EnableHistorical:  ko.Bool("chain.historical"),
		StartBlock:        uint64(ko.MustInt64("chain.start_block")),
		BatchSize:         ko.MustInt("chain.batch_size"),
		BatchQueue:        &batchQueue,
		BlocksQueue:       &blocksQueue,
		Chain:             chain,
		Logg:              lo,
		Stats:             stats,
		DB:                db,
	})
	if err != nil {
		lo.Error("could not initialize chain syncer", "error", err)
		os.Exit(1)
	}

	cache, err := cache.New(cache.CacheOpts{
		Logg:       lo,
		Chain:      chain,
		Registries: ko.MustStrings("bootstrap.ge_registries"),
		Blacklist:  ko.MustStrings("bootstrap.blacklist"),
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

	blockProcessor := processor.NewProcessor(processor.ProcessorOpts{
		Chain:       chain,
		BlocksQueue: &blocksQueue,
		Logg:        lo,
		Stats:       stats,
		DB:          db,
		Cache:       cache,
		Pub:         jetStreamPub,
	})

	if ko.Bool("chain.historical") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := chainSyncer.BootstrapHistoricalSyncer(); err != nil {
				lo.Error("could not bootstrap historical syncer", "error", err)
				os.Exit(1)
			}

			chainSyncer.StartHistoricalSyncer()
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		chainSyncer.StartRealtime()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		blockProcessor.Start()
	}()

	<-ctx.Done()
	lo.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		blockProcessor.Stop()
		jetStreamPub.Close()
		chainSyncer.StopHistoricalSyncer()
		chainSyncer.StopRealtime()
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