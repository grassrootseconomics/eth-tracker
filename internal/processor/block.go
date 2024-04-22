package processor

import (
	"context"
	"fmt"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/internal/handler"
)

func (p *Processor) processBlock(ctx context.Context, block types.Block) error {
	blockNumber := block.NumberU64()

	txs, err := p.chain.GetTransactions(ctx, block)
	if err != nil {
		return err
	}

	receiptsResp, err := p.chain.GetReceipts(ctx, block)
	if err != nil {
		return err
	}

	if len(receiptsResp) != len(txs) {
		return fmt.Errorf("block txs receipts len mismatch %d", blockNumber)
	}

	for i, receipt := range receiptsResp {
		if receipt.Status > 0 {
			for _, log := range receipt.Logs {
				if p.cache.Exists(log.Address.Hex()) {
					msg := handler.LogMessage{
						Log:       log,
						BlockTime: block.Time(),
					}

					if err := p.handleLogs(ctx, msg); err != nil {
						p.logg.Error("handler error", "handler_type", "log", "handler_name", "error", err)
					}
				}
			}
		} else {
			if txs[i].To() != nil && p.cache.Exists(txs[i].To().Hex()) {
				from, err := types.Sender(types.LatestSignerForChainID(txs[i].ChainId()), &txs[i])
				if err != nil {
					p.logg.Error("handler error", "handler_type", "revert", "error", err)
				}

				revertReason, err := p.chain.GetRevertReason(ctx, receipt.TxHash, receipt.BlockNumber)
				if err != nil {
					p.logg.Error("handler error", "handler_type", "revert", "error", err)
				}

				msg := handler.RevertMessage{
					From:            from.Hex(),
					RevertReason:    revertReason,
					InputData:       common.Bytes2Hex(txs[i].Data()),
					Block:           blockNumber,
					ContractAddress: txs[i].To().Hex(),
					Timestamp:       block.Time(),
					TxHash:          receipt.TxHash.Hex(),
				}

				if err := p.handleRevert(ctx, msg); err != nil {
					p.logg.Error("handler error", "handler_type", "revert", "error", err)
				}
			}
		}
	}

	if err := p.db.SetValue(blockNumber); err != nil {
		return err
	}
	p.logg.Debug("successfully processed block", "block", blockNumber)

	return nil
}

func (p *Processor) handleLogs(ctx context.Context, msg handler.LogMessage) error {
	for _, handler := range p.handlers {
		if err := handler.HandleLog(ctx, msg, p.emitter); err != nil {
			return fmt.Errorf("handler: %s err: %v", handler.Name(), err)
		}
	}

	return nil
}

func (p *Processor) handleRevert(ctx context.Context, msg handler.RevertMessage) error {
	for _, handler := range p.handlers {
		if err := handler.HandleRevert(ctx, msg, p.emitter); err != nil {
			return fmt.Errorf("handler: %s err: %v", handler.Name(), err)
		}
	}

	return nil
}
