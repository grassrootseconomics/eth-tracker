package chain

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/w3-celo/module/eth"
	"github.com/grassrootseconomics/w3-celo/w3types"
)

func (c *Chain) GetTransaction(ctx context.Context, txHash common.Hash) (types.Transaction, error) {
	var transaction types.Transaction
	if err := c.Provider.Client.CallCtx(ctx, eth.Tx(txHash).Returns(&transaction)); err != nil {
		return transaction, err
	}

	return transaction, nil
}

func (c *Chain) GetReceipts(ctx context.Context, block types.Block) ([]types.Receipt, error) {
	txCount := len(block.Transactions())

	calls := make([]w3types.RPCCaller, txCount)
	receipts := make([]types.Receipt, txCount)

	for i, tx := range block.Transactions() {
		calls[i] = eth.TxReceipt(tx.Hash()).Returns(&receipts[i])
	}

	if err := c.Provider.Client.CallCtx(ctx, calls...); err != nil {
		return nil, err
	}

	return receipts, nil
}

func (c *Chain) GetBlockReceipts(ctx context.Context, blockNumber *big.Int) (types.Receipts, error) {
	return nil, nil
}
