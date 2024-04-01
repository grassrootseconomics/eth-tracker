package processor

import (
	"context"
	"log/slog"

	"github.com/alitto/pond"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/ef-ds/deque/v2"
	"github.com/grassrootseconomics/celo-events/internal/chain"
	"github.com/grassrootseconomics/celo-events/internal/pool"
	"github.com/grassrootseconomics/celo-events/internal/stats"
)

type (
	ProcessorOpts struct {
		Chain       *chain.Chain
		BlocksQueue *deque.Deque[types.Block]
		Logg        *slog.Logger
		Stats       *stats.Stats
	}

	Processor struct {
		chain       *chain.Chain
		pool        *pond.WorkerPool
		blocksQueue *deque.Deque[types.Block]
		logg        *slog.Logger
		stats       *stats.Stats
	}
)

func NewProcessor(o ProcessorOpts) *Processor {
	return &Processor{
		chain:       o.Chain,
		pool:        pool.NewPool(o.Logg),
		blocksQueue: o.BlocksQueue,
		logg:        o.Logg,
		stats:       o.Stats,
	}
}

func (p *Processor) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			p.logg.Info("block processor shutting down")
			return nil
		default:
			for p.blocksQueue.Len() > 0 {
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

	transactionsResp, err := p.chain.GetTransactions(ctx, block)
	if err != nil {
		return err
	}

	receiptsResp, err := p.chain.GetReceipts(ctx, block)
	if err != nil {
		return err
	}

	for i, receipt := range receiptsResp {
		if receipt.Status < 1 {
			p.logg.Warn("reverted receipt", "tx_hash", receipt.TxHash, "input_data", common.Bytes2Hex(transactionsResp[i].Data()))
		}
		p.logg.Info("successful receipt", "tx_hash", receipt.TxHash, "status", receipt.Status)
	}

	return nil
}
