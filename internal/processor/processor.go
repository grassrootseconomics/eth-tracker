package processor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/handler"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
)

type (
	ProcessorOpts struct {
		Cache cache.Cache
		DB    db.DB
		Chain chain.Chain
		Pub   pub.Pub
		Logg  *slog.Logger
		Stats *stats.Stats
	}

	Processor struct {
		cache           cache.Cache
		db              db.DB
		chain           chain.Chain
		handlerPipeline handler.HandlerPipeline
		logg            *slog.Logger
		stats           *stats.Stats
	}
)

func NewProcessor(o ProcessorOpts) *Processor {
	return &Processor{
		cache:           o.Cache,
		db:              o.DB,
		handlerPipeline: handler.New(o.Pub, o.Cache),
		chain:           o.Chain,
		logg:            o.Logg,
		stats:           o.Stats,
	}
}

func (p *Processor) ProcessBlock(ctx context.Context, blockNumber uint64) error {
	block, err := p.chain.GetBlock(ctx, blockNumber)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("block %d error: %v", blockNumber, err)
	}

	receiptsResp, err := p.chain.GetReceipts(ctx, block)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("receipts fetch error: block %d: %v", blockNumber, err)
	}

	for _, receipt := range receiptsResp {
		if receipt.Status > 0 {
			for _, log := range receipt.Logs {
				if p.cache.Exists(log.Address.Hex()) {
					msg := handler.LogMessage{
						Log:       log,
						Timestamp: block.Time(),
					}

					if err := p.handleLog(ctx, msg); err != nil && !errors.Is(err, context.Canceled) {
						return fmt.Errorf("handle logs error: block %d: %v", blockNumber, err)
					}
				}
			}
		} else if p.isTrieAvailable(blockNumber) {
			tx, err := p.chain.GetTransaction(ctx, receipt.TxHash)
			if err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("get transaction error: tx %s: %v", receipt.TxHash.Hex(), err)
			}

			if tx.To() == nil {
				return nil
			}

			if p.cache.Exists(tx.To().Hex()) {
				from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
				if err != nil {
					return fmt.Errorf("transaction decode error: tx %s: %v", receipt.TxHash.Hex(), err)
				}

				revertReason, err := p.chain.GetRevertReason(ctx, receipt.TxHash, receipt.BlockNumber)
				if err != nil && !errors.Is(err, context.Canceled) {
					return fmt.Errorf("get revert reason error: tx %s: %v", receipt.TxHash.Hex(), err)
				}

				msg := handler.RevertMessage{
					From:            from.Hex(),
					RevertReason:    revertReason,
					InputData:       common.Bytes2Hex(tx.Data()),
					Block:           blockNumber,
					ContractAddress: tx.To().Hex(),
					Timestamp:       block.Time(),
					TxHash:          receipt.TxHash.Hex(),
				}

				if err := p.handleRevert(ctx, msg); err != nil && !errors.Is(err, context.Canceled) {
					return fmt.Errorf("handle revert error: tx %s: %v", receipt.TxHash.Hex(), err)
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

func (p *Processor) isTrieAvailable(blockNumber uint64) bool {
	available := p.chain.IsArchiveNode() || p.stats.GetLatestBlock()-blockNumber <= 128
	if !available {
		p.logg.Warn("skipping block due to potentially missing trie", "block_number", blockNumber)
	}
	return available
}

func (p *Processor) handleLog(ctx context.Context, msg handler.LogMessage) error {
	for _, handler := range p.handlerPipeline {
		if err := handler.HandleLog(ctx, msg); err != nil {
			return fmt.Errorf("log handler: %s err: %v", handler.Name(), err)
		}
	}

	return nil
}

func (p *Processor) handleRevert(ctx context.Context, msg handler.RevertMessage) error {
	for _, handler := range p.handlerPipeline {
		if err := handler.HandleRevert(ctx, msg); err != nil {
			return fmt.Errorf("revert handler: %s err: %v", handler.Name(), err)
		}
	}

	return nil
}
