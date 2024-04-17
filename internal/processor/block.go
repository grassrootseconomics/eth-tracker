package processor

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
)

func (p *Processor) processBlock(ctx context.Context, block types.Block) error {
	blockNumber := block.NumberU64()

	txs, err := p.chain.GetTransactions(ctx, block)
	p.logg.Debug("successfully fetched transactions", "txs", len(txs))
	if err != nil {
		return err
	}

	receiptsResp, err := p.chain.GetReceipts(ctx, block)
	p.logg.Debug("successfully fetched receipts", "receipts", len(txs))
	if err != nil {
		return err
	}

	for i, receipt := range receiptsResp {
		if receipt.Status > 0 {
			for _, log := range receipt.Logs {
				if err := p.handleLogs(ctx, log); err != nil {
					p.logg.Error("hanlder error", "error", err)
				}
			}
		} else {
			revertReason, _ := p.chain.GetRevertReason(ctx, receipt.TxHash, receipt.BlockNumber)
			p.logg.Debug("tx reverted", "hash", receipt.TxHash, "revert_reason", revertReason, "input_data", common.Bytes2Hex(txs[i].Data()))

		}
	}

	if err := p.db.SetValue(blockNumber); err != nil {
		return err
	}
	p.logg.Debug("successfully processed block", "block", blockNumber)

	return nil
}

func (p *Processor) handleLogs(ctx context.Context, log *types.Log) error {
	defaultEmitter := emitter.New(emitter.EmitterOpts{
		Logg: p.logg,
	})

	for _, handler := range p.handlers {
		if err := handler.Handle(ctx, log, defaultEmitter); err != nil {
			return err
		}
	}
	return nil
}
