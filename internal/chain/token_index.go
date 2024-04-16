package chain

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/w3-celo"
	"github.com/grassrootseconomics/w3-celo/module/eth"
	"github.com/grassrootseconomics/w3-celo/w3types"
)

var (
	entryCountFunc = w3.MustNewFunc("entryCount()", "uint256")
	entrySig       = w3.MustNewFunc("entry(uint256 _idx)", "address")
)

func (c *Chain) GetAllTokensFromTokenIndex(ctx context.Context, tokenIndex common.Address) ([]common.Address, error) {
	var (
		tokenIndexEntryCount big.Int
	)

	if err := c.provider.Client.CallCtx(
		ctx,
		eth.CallFunc(tokenIndex, entryCountFunc).Returns(&tokenIndexEntryCount),
	); err != nil {
		return nil, err
	}

	calls := make([]w3types.RPCCaller, tokenIndexEntryCount.Int64())
	tokenAddresses := make([]common.Address, tokenIndexEntryCount.Int64())

	for i := 0; i < int(tokenIndexEntryCount.Int64()); i++ {
		calls[i] = eth.CallFunc(tokenIndex, entrySig, new(big.Int).SetInt64(int64(i))).Returns(&tokenAddresses[i])
	}

	if err := c.provider.Client.CallCtx(ctx, calls...); err != nil {
		return nil, err
	}

	return tokenAddresses, nil
}
