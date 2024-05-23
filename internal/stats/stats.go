package stats

import (
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/alitto/pond"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
)

type (
	StatsOpts struct {
		Cache cache.Cache
		Logg  *slog.Logger
		Pool  *pond.WorkerPool
	}

	Stats struct {
		cache  cache.Cache
		logg   *slog.Logger
		pool   *pond.WorkerPool
		stopCh chan struct{}

		latestBlock atomic.Uint64
	}
)

const statsPrinterInterval = 5 * time.Second

func New(o StatsOpts) *Stats {
	return &Stats{
		cache:  o.Cache,
		logg:   o.Logg,
		pool:   o.Pool,
		stopCh: make(chan struct{}),
	}
}

func (s *Stats) SetLatestBlock(v uint64) {
	s.latestBlock.Store(v)
}

func (s *Stats) GetLatestBlock() uint64 {
	return s.latestBlock.Load()
}

func (s *Stats) Stop() {
	s.stopCh <- struct{}{}
}

func (s *Stats) APIStatsResponse() map[string]interface{} {
	return map[string]interface{}{
		"latestBlock":       s.GetLatestBlock(),
		"poolQueueSize":     s.pool.WaitingTasks(),
		"poolActiveWorkers": s.pool.RunningWorkers(),
		"cacheSize":         s.cache.Size(),
	}
}

func (s *Stats) StartStatsPrinter() {
	ticker := time.NewTicker(statsPrinterInterval)

	for {
		select {
		case <-s.stopCh:
			s.logg.Debug("stats shutting down")
			return
		case <-ticker.C:
			s.logg.Info("block stats",
				"latest_block", s.GetLatestBlock(),
				"pool_queue_size", s.pool.WaitingTasks(),
				"pool_active_workers", s.pool.RunningWorkers(),
				"cache_size", s.cache.Size(),
			)
		}
	}
}
