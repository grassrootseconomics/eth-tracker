package processor

import (
	"context"
	"log/slog"

	"github.com/alitto/pond"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/ef-ds/deque/v2"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
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
	}
)

func NewProcessor(o ProcessorOpts) *Processor {
	return &Processor{
		chain:       o.Chain,
		pool:        pool.NewPool(o.Logg),
		blocksQueue: o.BlocksQueue,
		logg:        o.Logg,
		stats:       o.Stats,
		db:          o.DB,
	}
}

func (p *Processor) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			p.logg.Info("block processor shutting down")
			p.Stop()
			return
		default:
			if p.blocksQueue.Len() > 0 {
				v, _ := p.blocksQueue.PopFront()
				p.pool.Submit(func() {
					if err := p.processBlock(v); err != nil {
						p.logg.Info("block processor error", "block", v.NumberU64(), "error", err)
					}
				})
			}
		}
	}

}
func (p *Processor) Stop() {
	p.pool.StopAndWait()
}

func (p *Processor) processBlock(block types.Block) error {
	ctx := context.Background()
	blockNumber := block.NumberU64()

	_, err := p.chain.GetTransactions(ctx, block)
	if err != nil {
		return err
	}

	receiptsResp, err := p.chain.GetReceipts(ctx, block)
	if err != nil {
		return err
	}

	for _, receipt := range receiptsResp {
		if receipt.Status < 1 {
			//
		}
	}

	if err := p.db.SetValue(blockNumber); err != nil {
		return err
	}
	p.logg.Debug("successfully processed block", "block", blockNumber)

	return nil
}
