package syncer

import (
	"context"
	"fmt"
)

func (s *Syncer) BootstrapHistoricalSyncer() {
	// logg here
	for i, e := s.db.NextSet(0); e; i, e = s.db.NextSet(i + 1) {
		if i > 0 {
			s.batchQueue.PushBack(uint64(i))
		}
	}
}

func (s *Syncer) StartHistoricalSyncer(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			s.logg.Info("historical syncer shutting down")
			return nil
		default:
			for s.batchQueue.Len() > 0 {
				var (
					currentIterLen = s.batchQueue.Len()
					batch          []uint64
				)

				if currentIterLen < blockBatchSize {
					batch = make([]uint64, currentIterLen)
					for i := 0; i < currentIterLen; i++ {
						v, _ := s.batchQueue.PopFront()
						batch[i] = v
					}
				} else {
					batch = make([]uint64, blockBatchSize)
					for i := 0; i < blockBatchSize; i++ {
						v, _ := s.batchQueue.PopFront()
						batch[i] = v
					}
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
