package chain

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/w3-celo/module/eth"
	"github.com/grassrootseconomics/w3-celo/w3types"
)

func (c *Chain) GetBlocks(ctx context.Context, blockNumbers []uint64) ([]types.Block, error) {
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

func (c *Chain) GetBlock(ctx context.Context, blockNumber uint64) (types.Block, error) {
	var (
		block types.Block
	)

	blockCall := eth.BlockByNumber(new(big.Int).SetUint64(blockNumber)).Returns(&block)

	if err := c.provider.Client.CallCtx(ctx, blockCall); err != nil {
		return block, err
	}

	return block, nil
}
