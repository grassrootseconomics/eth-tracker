package chain

import (
	"context"

	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/w3-celo/module/eth"
	"github.com/grassrootseconomics/w3-celo/w3types"
)

func (c *Chain) GetTransactions(ctx context.Context, block types.Block) ([]types.Transaction, error) {
	txCount := len(block.Transactions())

	calls := make([]w3types.RPCCaller, txCount)
	transactions := make([]types.Transaction, txCount)

	for i, tx := range block.Transactions() {
		calls[i] = eth.Tx(tx.Hash()).Returns(&transactions[i])
	}

	if err := c.provider.Client.CallCtx(ctx, calls...); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (c *Chain) GetReceipts(ctx context.Context, block types.Block) ([]types.Receipt, error) {
	txCount := len(block.Transactions())

	calls := make([]w3types.RPCCaller, txCount)
	receipts := make([]types.Receipt, txCount)

	for i, tx := range block.Transactions() {
		calls[i] = eth.TxReceipt(tx.Hash()).Returns(&receipts[i])
	}

	if err := c.provider.Client.CallCtx(ctx, calls...); err != nil {
		return nil, err
	}

	return receipts, nil
}
