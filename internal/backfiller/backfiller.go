package backfiller

import (
	"context"
	"log/slog"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/processor"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
	"github.com/grassrootseconomics/celo-tracker/pkg/chain"
)

type (
	BackfillerOpts struct {
		BlockWorker    *workerpool.WorkerPool
		BlockProcessor *processor.Processor
		Chain          *chain.Chain
		DB             *db.DB
		Logg           *slog.Logger
		Stats          *stats.Stats
	}

	Backfiller struct {
		blockWorker    *workerpool.WorkerPool
		blockProcessor *processor.Processor
		chain          *chain.Chain
		db             *db.DB
		logg           *slog.Logger
		quit           chan struct{}
		stats          *stats.Stats
	}
)

const (
	blockBatchSize   = 50
	verifierInterval = 20 * time.Second
)

func New(o BackfillerOpts) *Backfiller {
	return &Backfiller{
		blockWorker:    o.BlockWorker,
		blockProcessor: o.BlockProcessor,
		chain:          o.Chain,
		db:             o.DB,
		logg:           o.Logg,
		quit:           make(chan struct{}),
		stats:          o.Stats,
	}
}

func (b *Backfiller) Start() {
	ticker := time.NewTicker(verifierInterval)

	for {
		select {
		case <-b.quit:
			b.logg.Info("verifier shutting down")
			return
		case <-ticker.C:
			if b.blockWorker.WaitingQueueSize() <= 1 {
				if err := b.Run(true); err != nil {
					b.logg.Error("verifier run error", "err", err)
				}
			}
		}
	}
}

func (b *Backfiller) Stop() {
	b.quit <- struct{}{}
}

func (b *Backfiller) Run(skipLatest bool) error {
	lower, err := b.db.GetLowerBound()
	if err != nil {
		return err
	}
	upper, err := b.db.GetUpperBound()
	if err != nil {
		return err
	}

	if skipLatest {
		upper--
	}

	missingBlocks, err := b.db.GetMissingValuesBitSet(lower, upper)
	if err != nil {
		return err
	}
	missingBlocksCount := missingBlocks.Count()
	if missingBlocksCount > 0 {
		b.logg.Info("found missing blocks", "skip_latest", skipLatest, "missing_blocks_count", missingBlocksCount)
		buffer := make([]uint, missingBlocksCount)
		missingBlocks.NextSetMany(0, buffer)
		defer missingBlocks.ClearAll()

		for i := 0; i < int(missingBlocksCount); i += blockBatchSize {
			end := i + blockBatchSize
			if end > int(missingBlocksCount) {
				end = int(missingBlocksCount)
			}
			batch := make([]uint64, end-i)
			for j := i; j < end; j++ {
				batch[j-i] = uint64(buffer[j])
			}

			blocks, err := b.chain.GetBlocks(context.Background(), batch)
			if err != nil {
				return err
			}

			for _, block := range blocks {
				b.blockWorker.Submit(func() {
					if err := b.blockProcessor.ProcessBlock(context.Background(), block); err != nil {
						b.logg.Error("block processor error", "source", "backfiller", "block", block.NumberU64(), "error", err)
					}
				})
			}
		}
	}

	return nil
}
