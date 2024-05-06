package verifier

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/processor"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
	"github.com/grassrootseconomics/celo-tracker/pkg/chain"
)

type (
	VerifierOpts struct {
		BlockWorker    *workerpool.WorkerPool
		BlockProcessor *processor.Processor
		Chain          *chain.Chain
		DB             *db.DB
		Logg           *slog.Logger
		Stats          *stats.Stats
	}

	Verifier struct {
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
	blockBatchSize   = 25
	verifierInterval = 1 * time.Minute
)

func New(o VerifierOpts) *Verifier {
	return &Verifier{
		blockWorker: o.BlockWorker,
		chain:       o.Chain,
		db:          o.DB,
		logg:        o.Logg,
		quit:        make(chan struct{}),
		stats:       o.Stats,
	}
}

func (v *Verifier) Start() {
	ticker := time.NewTicker(verifierInterval)

	for {
		select {
		case <-v.quit:
			v.logg.Info("janitor: shutdown signal received")
			return
		case <-ticker.C:
			batch, err := v.getMissingBlocks()
			if err != nil {
				v.logg.Error("verifier error getting missing blocks", "err", err)
			}

			if batch != nil {
				v.logg.Info("verifier found missing block gap requeuing missing blocks")
				blocks, err := v.chain.GetBlocks(context.Background(), batch)
				if err != nil {
					v.logg.Error("batch blocks fetcher error", "error", "block_range", fmt.Sprintf("%d-%d", batch[0], batch[len(batch)-1]), "error", err)
				}

				for _, block := range blocks {
					v.blockWorker.Submit(func() {
						if err := v.blockProcessor.ProcessBlock(context.Background(), block); err != nil {
							v.logg.Error("block processor error", "source", "verifier", "block", block.NumberU64(), "error", err)
						}
					})
				}
			} else {
				v.logg.Debug("verifier found no missing blocks running db compactor")
				if err := v.db.Cleanup(); err != nil {
					v.logg.Error("verifier compactor error", "error", err)
				}
			}
		}
	}
}

func (v *Verifier) Stop() {
	// TODO: Run with sync.Once
	v.quit <- struct{}{}
}

func (v *Verifier) getMissingBlocks() ([]uint64, error) {
	lower, err := v.db.GetLowerBound()
	if err != nil {
		return nil, err
	}

	upper, err := v.db.GetUpperBound()
	if err != nil {
		return nil, err
	}

	missingBlocks, err := v.db.GetMissingValuesBitSet(lower, upper-1)
	if err != nil {
		return nil, err
	}
	missingBlocksCount := missingBlocks.Count()

	if missingBlocksCount > 0 {

		buffer := make([]uint, missingBlocksCount)
		missingBlocks.NextSetMany(0, buffer)

		for i := 0; i < int(missingBlocksCount); i += blockBatchSize {
			end := i + blockBatchSize
			if end > int(missingBlocksCount) {
				end = int(missingBlocksCount)
			}
			batch := make([]uint64, end-i)
			for j := i; j < end; j++ {
				batch[j-i] = uint64(buffer[j])
			}

			return batch, nil
		}
	}

	return nil, nil
}
