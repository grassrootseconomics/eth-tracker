package verifier

import (
	"context"
	"fmt"
	"log/slog"

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
	blockBatchSize = 25
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

func (v *Verifier) Start()

func (v *Verifier) getMissingBlocks() error {
	lower, err := v.db.GetLowerBound()
	if err != nil {
		return err
	}

	upper, err := v.db.GetUpperBound()
	if err != nil {
		return err
	}

	missingBlocks, err := v.db.GetMissingValuesBitSet(lower, upper-1)
	if err != nil {
		return err
	}
	missingBlocksCount := missingBlocks.Count()

	if missingBlocksCount > 0 {
		v.logg.Info("verifier found block gap", "missing_blocks_count", missingBlocksCount, "lower_bound", lower, "upper_bound", upper)

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

			v.processMissingBlocksBatch(batch)
		}
	} else {
		v.logg.Debug("verifier running db compactor")
		if err := v.db.Cleanup(); err != nil {
			v.logg.Error("verifier compactor error", "error", err)
		}
	}

	return nil
}

func (v *Verifier) processMissingBlocksBatch(batch []uint64) {
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
}
