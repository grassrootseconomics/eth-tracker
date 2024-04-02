package stats

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

const (
	statsPrinterInterval = 5 * time.Second
)

type (
	Stats struct {
		logg *slog.Logger

		latestBlock           atomic.Uint64
		latestCheckpointBlock atomic.Uint64
		missingCount          atomic.Uint32
		historicalQueueSize   atomic.Uint32
	}
)

func New(logg *slog.Logger) *Stats {
	return &Stats{
		logg: logg,
	}
}

func (s *Stats) SetLatestBlock(v uint64) {
	s.latestBlock.Store(v)
}

func (s *Stats) GetLatestBlock() uint64 {
	return s.latestBlock.Load()
}

func (s *Stats) SetMissingCount(v uint32) {
	s.missingCount.Store(v)
}

func (s *Stats) GetMissingCount() uint32 {
	return s.missingCount.Load()
}

func (s *Stats) SetLatestCheckpointBlock(v uint64) {
	s.latestCheckpointBlock.Store(v)
}

func (s *Stats) GetLatestChckpointBlock() uint64 {
	return s.latestCheckpointBlock.Load()
}

func (s *Stats) SetHistoricalQueueSize(v uint32) {
	s.historicalQueueSize.Store(v)
}

func (s *Stats) GetHistoricalQueueSize() uint32 {
	return s.historicalQueueSize.Load()
}

func (s *Stats) StartStatsPrinter(ctx context.Context) error {

	ticker := time.NewTicker(1 * statsPrinterInterval)
	for {
		select {
		case <-ctx.Done():
			s.logg.Debug("stats shutting down")
			return nil
		case <-ticker.C:
			s.logg.Info("tracker stats", "latest_block", s.GetLatestBlock(), "missing_count", s.GetMissingCount(), "latest_checkpoint", s.GetLatestChckpointBlock())
		}
	}
}
