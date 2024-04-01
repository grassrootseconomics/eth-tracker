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
	"github.com/grassrootseconomics/celo-events/internal/chain"
	"github.com/grassrootseconomics/celo-events/internal/processor"
	"github.com/grassrootseconomics/celo-events/internal/stats"
	"github.com/grassrootseconomics/celo-events/internal/syncer"
	"github.com/knadh/koanf/v2"
)

const defaultGracefulShutdownPeriod = time.Second * 10

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
	stats := stats.NewStats(lo)

	chain, err := chain.NewChainProvider(chain.ChainOpts{
		RPCEndpoint: ko.MustString("chain.rpc_endpoint"),
		TestNet:     ko.Bool("chain.testnet"),
		Logg:        lo,
	})
	if err != nil {
		lo.Error("could not initialize chain client", "error", err)
	}

	chainSyncer, err := syncer.NewSyncer(syncer.SyncerOpts{
		WebSocketEndpoint: ko.MustString("chain.ws_endpoint"),
		BatchQueue:        &batchQueue,
		BlocksQueue:       &blocksQueue,
		Chain:             chain,
		Logg:              lo,
		Stats:             stats,
	})
	if err != nil {
		lo.Error("could not initialize chain syncer", "error", err)
	}
	// chainSyncer.BootstrapHistoricalSyncer()

	blockProcessor := processor.NewProcessor(processor.ProcessorOpts{
		Chain:       chain,
		BlocksQueue: &blocksQueue,
		Logg:        lo,
		Stats:       stats,
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		chainSyncer.StartRealtimeSyncer(ctx)
	}()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	chainSyncer.StartHistoricalSyncer(ctx)
	// }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		blockProcessor.Start(ctx)
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		blockProcessor.Stop()
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
