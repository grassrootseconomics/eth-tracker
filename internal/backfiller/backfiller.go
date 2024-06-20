package backfiller

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/queue"
)

type (
	BackfillerOpts struct {
		MaxQueueSize int
		DB           db.DB
		Logg         *slog.Logger
		Queue        *queue.Queue
	}

	backfiller struct {
		maxQueueSize int
		db           db.DB
		logg         *slog.Logger
		queue        *queue.Queue
		stopCh       chan struct{}
		ticker       *time.Ticker
	}
)

const verifierInterval = 20 * time.Second

func New(o BackfillerOpts) *backfiller {
	return &backfiller{
		db:     o.DB,
		logg:   o.Logg,
		queue:  o.Queue,
		stopCh: make(chan struct{}),
		ticker: time.NewTicker(verifierInterval),
	}
}

func (b *backfiller) Stop() {
	b.ticker.Stop()
	b.stopCh <- struct{}{}
}

func (b *backfiller) Start() {
	for {
		select {
		case <-b.stopCh:
			b.logg.Info("verifier shutting down")
			b.ticker.Stop()
			return
		case <-b.ticker.C:
			if b.queue.Size() <= 1 {
				if err := b.Run(true); err != nil {
					b.logg.Error("verifier tick run error", "err", err)
				}
				b.logg.Debug("verifier successful run", "queue_size", b.queue.Size())
			} else {
				b.logg.Debug("skipping verifier run")
			}
		}
	}
}

func (b *backfiller) Run(skipLatest bool) error {
	lower, err := b.db.GetLowerBound()
	if err != nil {
		return fmt.Errorf("verifier could not get lower bound from db: err %v", err)
	}
	upper, err := b.db.GetUpperBound()
	if err != nil {
		return fmt.Errorf("verifier could not get upper bound from db: err %v", err)
	}

	if skipLatest {
		upper--
	}

	missingBlocks, err := b.db.GetMissingValuesBitSet(lower, upper)
	if err != nil {
		return fmt.Errorf("verifier could not get missing values bitset: err %v", err)
	}
	missingBlocksCount := missingBlocks.Count()

	if missingBlocksCount > 0 {
		if missingBlocksCount >= uint(b.maxQueueSize) {
			b.logg.Warn("large number of blocks missing this may result in degraded RPC performance set FORCE_BACKFILL=* to continue", "missing_blocks", missingBlocksCount)
			_, ok := os.LookupEnv("FORCE_BACKFILL")
			if !ok {
				os.Exit(0)
			}
		}
		b.logg.Info("bootstrapping queue with missing blocks")

		b.logg.Info("found missing blocks", "skip_latest", skipLatest, "missing_blocks_count", missingBlocksCount)
		buffer := make([]uint, missingBlocksCount)
		missingBlocks.NextSetMany(0, buffer)
		defer missingBlocks.ClearAll()

		for _, block := range buffer {
			b.queue.Push(uint64(block))
		}
	}

	return nil
}
