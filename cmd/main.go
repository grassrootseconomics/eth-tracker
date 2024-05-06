package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/pool"
	"github.com/grassrootseconomics/celo-tracker/internal/processor"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
	"github.com/grassrootseconomics/celo-tracker/internal/syncer"
	"github.com/grassrootseconomics/celo-tracker/internal/verifier"
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
	var wg sync.WaitGroup
	ctx, stop := notifyShutdown()

	/*
		Dependency Order
		----------------
		- Stats
		- Chain
		- DB
		- Cache
		- JetStream Pub
		- Worker Pool
		- Block Processor
		- Chain Syncer
		- Verifier

	*/
	stats := stats.New(lo)

	chain, err := chain.New(chain.ChainOpts{
		Logg:        lo,
		RPCEndpoint: ko.MustString("chain.rpc_endpoint"),
		TestNet:     ko.Bool("chain.testnet"),
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

	workerPool := pool.NewPool(pool.PoolOpts{
		PoolSize: runtime.NumCPU(),
	})

	blockProcessor := processor.NewProcessor(processor.ProcessorOpts{
		Cache: cache,
		Chain: chain,
		DB:    db,
		Logg:  lo,
		Pub:   jetStreamPub,
		Stats: stats,
	})

	chainSyncer, err := syncer.New(syncer.SyncerOpts{
		BlockWorker:       workerPool,
		BlockProcessor:    blockProcessor,
		Chain:             chain,
		DB:                db,
		Logg:              lo,
		Stats:             stats,
		WebSocketEndpoint: ko.MustString("chain.ws_endpoint"),
	})
	if err != nil {
		lo.Error("could not initialize chain syncer", "error", err)
		os.Exit(1)
	}

	verifier := verifier.New(verifier.VerifierOpts{
		BlockWorker:    workerPool,
		BlockProcessor: blockProcessor,
		Chain:          chain,
		DB:             db,
		Logg:           lo,
		Stats:          stats,
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		chainSyncer.Start()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		verifier.Start()
	}()

	<-ctx.Done()
	lo.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		workerPool.StopWait()
		chainSyncer.Stop()
		verifier.Stop()
		jetStreamPub.Close()
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
