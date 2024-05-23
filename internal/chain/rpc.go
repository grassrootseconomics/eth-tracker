package chain

import (
	"context"
	"math/big"
	"net/http"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/rpc"
	"github.com/grassrootseconomics/celoutils/v3"
	"github.com/grassrootseconomics/w3-celo"
	"github.com/grassrootseconomics/w3-celo/module/eth"
	"github.com/grassrootseconomics/w3-celo/w3types"
)

type (
	RPCOpts struct {
		RPCEndpoint   string
		ChainID       int64
		IsArchiveNode bool
	}

	RPC struct {
		provider      *celoutils.Provider
		isArchiveNode bool
	}
)

func NewRPCFetcher(o RPCOpts) (Chain, error) {
	customRPCClient, err := lowTimeoutRPCClient(o.RPCEndpoint)
	if err != nil {
		return nil, err
	}

	chainProvider := celoutils.NewProvider(
		o.RPCEndpoint,
		o.ChainID,
		celoutils.WithClient(customRPCClient),
	)

	return &RPC{
		provider:      chainProvider,
		isArchiveNode: o.IsArchiveNode,
	}, nil
}

func lowTimeoutRPCClient(rpcEndpoint string) (*w3.Client, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	rpcClient, err := rpc.DialHTTPWithClient(
		rpcEndpoint,
		httpClient,
	)
	if err != nil {
		return nil, err
	}

	return w3.NewClient(rpcClient), nil
}

func (c *RPC) GetBlocks(ctx context.Context, blockNumbers []uint64) ([]types.Block, error) {
	blocksCount := len(blockNumbers)
	calls := make([]w3types.RPCCaller, blocksCount)
	blocks := make([]types.Block, blocksCount)

	for i, v := range blockNumbers {
		calls[i] = eth.BlockByNumber(new(big.Int).SetUint64(v)).Returns(&blocks[i])
	}

	if err := c.provider.Client.CallCtx(ctx, calls...); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (c *RPC) GetBlock(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	var block types.Block
	blockCall := eth.BlockByNumber(new(big.Int).SetUint64(blockNumber)).Returns(&block)

	if err := c.provider.Client.CallCtx(ctx, blockCall); err != nil {
		return nil, err
	}

	return &block, nil
}

func (c *RPC) GetLatestBlock(ctx context.Context) (uint64, error) {
	var latestBlock big.Int
	latestBlockCall := eth.BlockNumber().Returns(&latestBlock)

	if err := c.provider.Client.CallCtx(ctx, latestBlockCall); err != nil {
		return 0, err
	}

	return latestBlock.Uint64(), nil
}

func (c *RPC) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, error) {
	var transaction types.Transaction
	if err := c.provider.Client.CallCtx(ctx, eth.Tx(txHash).Returns(&transaction)); err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (c *RPC) GetReceipts(ctx context.Context, block *types.Block) ([]types.Receipt, error) {
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

func (c *RPC) GetRevertReason(ctx context.Context, txHash common.Hash, blockNumber *big.Int) (string, error) {
	return c.provider.SimulateRevertedTx(ctx, txHash, blockNumber)
}

func (c *RPC) Provider() *celoutils.Provider {
	return c.provider
}

func (c *RPC) IsArchiveNode() bool {
	return c.isArchiveNode
}
