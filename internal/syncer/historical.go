package syncer

import (
	"context"
	"fmt"
	"time"
)

const (
	emptyQueueIdelTime = 1 * time.Second
)

func (s *Syncer) BootstrapHistoricalSyncer() error {
	lower, err := s.db.GetLowerBound()
	if err != nil {
		return err
	}

	upper, err := s.db.GetUpperBound()
	if err != nil {
		return err
	}

	missingBlocks, err := s.db.GetMissingValuesBitSet(lower, upper)
	if err != nil {
		return err
	}
	missingBlocksCount := missingBlocks.Count()
	s.logg.Info("bootstrapping historical syncer", "missing_blocks", missingBlocksCount, "lower_bound", lower, "upper_bound", upper)

	buffer := make([]uint, missingBlocksCount)
	missingBlocks.NextSetMany(0, buffer)
	for _, v := range buffer {
		s.batchQueue.PushFront(uint64(v))
	}

	return nil
}

func (s *Syncer) StartHistoricalSyncer() error {
	s.logg.Info("starting historical syncer", "batch_size", s.batchSize)
	for {
		select {
		case <-s.quit:
			s.logg.Info("historical syncer stopped")
			return nil
		default:
			if s.batchQueue.Len() > 0 {
				var (
					currentIterLen = s.batchQueue.Len()
				)

				if currentIterLen > s.batchSize {
					currentIterLen = s.batchSize
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
			} else {
				time.Sleep(emptyQueueIdelTime)
				s.logg.Debug("historical batcher queue empty slept for 2 seconds")
			}
		}
	}
}

func (s *Syncer) StopHistoricalSyncer() {
	s.logg.Info("signaling historical syncer shutdown")
	s.quit <- struct{}{}
}
