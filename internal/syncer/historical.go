package syncer

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

const (
	blockBatchSize = 100
)

func (s *Syncer) BootstrapHistoricalSyncer() error {
	v, err := s.db.GetLowerBound()
	if err != nil {
		if err == badger.ErrKeyNotFound {
			if err := s.db.SetLowerBound(s.initialLowerBound); err != nil {
				return err
			}
			v = s.initialLowerBound
		} else {
			return err
		}
	}

	latestBlock, err := s.chain.GetLatestBlock(context.Background())
	if err != nil {
		return err
	}
	if err := s.db.SetUpperBound(latestBlock); err != nil {
		return err
	}

	missingBlocks, err := s.db.GetMissingValuesBitSet(v, latestBlock)
	if err != nil {
		return err
	}
	missingBlocksCount := missingBlocks.Count()
	s.logg.Info("bootstrapping historical syncer", "missing_blocks", missingBlocksCount, "lower_bound", v, "upper_bound", latestBlock)

	buffer := make([]uint, missingBlocksCount)
	missingBlocks.NextSetMany(0, buffer)
	for _, v := range buffer {
		s.batchQueue.PushFront(uint64(v))
	}

	return nil
}

func (s *Syncer) StartHistoricalSyncer(ctx context.Context) error {
	s.logg.Info("starting historical syncer", "batch_size", blockBatchSize)
	for {
		select {
		case <-ctx.Done():
			s.logg.Info("historical syncer shutting down")
			return nil
		default:
			if s.batchQueue.Len() > 0 {
				var (
					currentIterLen = s.batchQueue.Len()
				)

				if currentIterLen > blockBatchSize {
					currentIterLen = blockBatchSize
				}
				batch := make([]uint64, currentIterLen)
				for i := 0; i < currentIterLen; i++ {
					v, _ := s.batchQueue.PopFront()
					batch[i] = v
				}

				blocks, err := s.chain.GetBlocks(context.Background(), batch)
				if err != nil {
					s.logg.Error("batch blocks fetcher error", "fetch_size", currentIterLen, "block_range", fmt.Sprintf("%d-%d", batch[0], batch[len(batch)-1]), "error", err)
				}

				for _, v := range blocks {
					s.blocksQueue.PushBack(v)
				}
			}
		}
	}
}
