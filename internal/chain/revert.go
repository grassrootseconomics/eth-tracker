package chain

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
)

func (c *Chain) GetRevertReason(ctx context.Context, txHash common.Hash, blockNumber *big.Int) (string, error) {
	return c.provider.SimulateRevertedTx(ctx, txHash, blockNumber)
}
