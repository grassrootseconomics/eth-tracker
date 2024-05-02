package processor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/handler"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
	"github.com/grassrootseconomics/celo-tracker/pkg/chain"
)

type (
	ProcessorOpts struct {
		Cache cache.Cache
		Chain *chain.Chain
		DB    *db.DB
		Logg  *slog.Logger
		Pub   pub.Pub
		Stats *stats.Stats
	}

	Processor struct {
		cache           cache.Cache
		chain           *chain.Chain
		db              *db.DB
		handlerPipeline handler.HandlerPipeline
		logg            *slog.Logger
		pub             pub.Pub
		quit            chan struct{}
		stats           *stats.Stats
	}
)

func NewProcessor(o ProcessorOpts) *Processor {
	return &Processor{
		cache:           o.Cache,
		chain:           o.Chain,
		db:              o.DB,
		handlerPipeline: handler.New(o.Cache),
		logg:            o.Logg,
		pub:             o.Pub,
		quit:            make(chan struct{}),
		stats:           o.Stats,
	}
}

func (p *Processor) ProcessBlock(ctx context.Context, block types.Block) error {
	receiptsResp, err := p.chain.GetReceipts(ctx, block)
	if err != nil {
		return err
	}

	for _, receipt := range receiptsResp {
		if receipt.Status > 0 {
			for _, log := range receipt.Logs {
				if p.cache.Exists(log.Address.Hex()) {
					msg := handler.LogMessage{
						Log:       log,
						Timestamp: block.Time(),
					}

					if err := p.handleLogs(ctx, msg); err != nil {
						return err
					}
				}
			}
		} else {
			tx, err := p.chain.GetTransaction(ctx, receipt.TxHash)
			if err != nil {
				return err
			}

			if tx.To() != nil && p.cache.Exists(tx.To().Hex()) {
				from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), &tx)
				if err != nil {
					return err
				}

				revertReason, err := p.chain.GetRevertReason(ctx, receipt.TxHash, receipt.BlockNumber)
				if err != nil {
					return err
				}

				msg := handler.RevertMessage{
					From:            from.Hex(),
					RevertReason:    revertReason,
					InputData:       common.Bytes2Hex(tx.Data()),
					Block:           block.NumberU64(),
					ContractAddress: tx.To().Hex(),
					Timestamp:       block.Time(),
					TxHash:          receipt.TxHash.Hex(),
				}

				if err := p.handleRevert(ctx, msg); err != nil {
					return err
				}
			}
		}
	}

	if err := p.db.SetValue(block.NumberU64()); err != nil {
		return err
	}
	p.logg.Debug("successfully processed block", "block", block.NumberU64())

	return nil
}

func (p *Processor) handleLogs(ctx context.Context, msg handler.LogMessage) error {
	for _, handler := range p.handlerPipeline {
		if err := handler.HandleLog(ctx, msg, p.pub); err != nil {
			return fmt.Errorf("log handler: %s err: %v", handler.Name(), err)
		}
	}

	return nil
}

func (p *Processor) handleRevert(ctx context.Context, msg handler.RevertMessage) error {
	for _, handler := range p.handlerPipeline {
		if err := handler.HandleRevert(ctx, msg, p.pub); err != nil {
			return fmt.Errorf("revert handler: %s err: %v", handler.Name(), err)
		}
	}

	return nil
}
