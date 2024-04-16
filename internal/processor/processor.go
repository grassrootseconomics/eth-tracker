package processor

import (
	"context"
	"log/slog"
	"time"

	"github.com/alitto/pond"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/ef-ds/deque/v2"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/handler"
	"github.com/grassrootseconomics/celo-tracker/internal/pool"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
)

type (
	ProcessorOpts struct {
		Chain       *chain.Chain
		BlocksQueue *deque.Deque[types.Block]
		Logg        *slog.Logger
		Stats       *stats.Stats
		DB          *db.DB
	}

	Processor struct {
		chain       *chain.Chain
		pool        *pond.WorkerPool
		blocksQueue *deque.Deque[types.Block]
		logg        *slog.Logger
		stats       *stats.Stats
		db          *db.DB
		quit        chan struct{}
		handlers    []handler.Handler
	}
)

const (
	emptyQueueIdleTime = 1 * time.Second
)

func NewProcessor(o ProcessorOpts) *Processor {
	return &Processor{
		chain:       o.Chain,
		pool:        pool.NewPool(o.Logg),
		blocksQueue: o.BlocksQueue,
		logg:        o.Logg,
		stats:       o.Stats,
		db:          o.DB,
		quit:        make(chan struct{}),
		handlers:    handler.New(),
	}
}

func (p *Processor) Start() {
	for {
		select {
		case <-p.quit:
			p.logg.Info("processor stopped, draining workerpool queue")
			p.pool.StopAndWait()
			if err := p.db.Close(); err != nil {
				p.logg.Info("error closing db", "error", err)
			}
			return
		default:
			if p.blocksQueue.Len() > 0 {
				v, _ := p.blocksQueue.PopFront()
				p.pool.Submit(func() {
					p.logg.Info("processing", "block", v.Number())
					if err := p.processBlock(context.Background(), v); err != nil {
						p.logg.Info("block processor error", "block", v.NumberU64(), "error", err)
					}
				})
			} else {
				time.Sleep(emptyQueueIdleTime)
				p.logg.Debug("queue empty slept for 1 second")
			}
		}
	}
}

func (p *Processor) Stop() {
	p.logg.Info("signaling processor shutdown")
	p.quit <- struct{}{}
}
