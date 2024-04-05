package chain

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/w3-celo"
	"github.com/grassrootseconomics/w3-celo/module/eth"
)

func (c *Chain) GetRevertReason(ctx context.Context, txHash common.Hash, blockNumber *big.Int) (string, error) {
	return c.provider.SimulateRevertedTx(ctx, txHash, blockNumber)
}

func (c *Chain) TestDecodeTransfer(ctx context.Context, logs []*types.Log) {
	signature := "Transfer(address indexed _from, address indexed _to, uint256 _value)"

	eventTransfer := w3.MustNewEvent(signature)

	for _, log := range logs {
		if log.Topics[0] == w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef") {
			var (
				from  common.Address
				to    common.Address
				value big.Int

				tokenSymbol   string
				tokenDecimals big.Int
			)

			if err := c.provider.Client.CallCtx(
				ctx,
				eth.CallFunc(log.Address, w3.MustNewFunc("symbol()", "string")).Returns(&tokenSymbol),
				eth.CallFunc(log.Address, w3.MustNewFunc("decimals()", "uint256")).Returns(&tokenDecimals),
			); err != nil {
				c.logg.Error("token details fetcher", "error", err)
			}

			if err := eventTransfer.DecodeArgs(log, &from, &to, &value); err != nil {
				c.logg.Error("event decoder", "error", err)
			}

			c.logg.Info("transfer event",
				"hash", log.TxHash,
				"token", tokenSymbol,
				"from", from,
				"to", to,
				"value", value.Uint64(),
			)
		}
	}

}
