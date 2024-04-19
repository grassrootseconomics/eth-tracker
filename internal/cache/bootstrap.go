package cache

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celoutils/v2"
	"github.com/grassrootseconomics/w3-celo"
	"github.com/grassrootseconomics/w3-celo/module/eth"
)

func bootstrapGESmartContracts(ctx context.Context, registries []string, chain *chain.Chain, cache Cache) (WatchableIndex, error) {
	var (
		watchableIndex = make(WatchableIndex)
	)

	for _, registry := range registries {
		registryMap, err := chain.Provider.RegistryMap(ctx, w3.A(registry))
		if err != nil {
			return nil, nil
		}

		for _, v := range registryMap {
			cache.Add(v.Hex())
		}

		if registryMap[celoutils.TokenIndex] != common.ZeroAddress {
			tokens, err := chain.GetAllTokensFromTokenIndex(ctx, registryMap[celoutils.TokenIndex])
			if err != nil {
				return nil, err
			}

			for _, token := range tokens {
				cache.Add(token.Hex())
			}

			watchableIndex[registryMap[celoutils.TokenIndex].Hex()] = true
		}

		if registryMap[celoutils.PoolIndex] != common.ZeroAddress {
			pools, err := chain.GetAllTokensFromTokenIndex(ctx, registryMap[celoutils.PoolIndex])
			if err != nil {
				return nil, err
			}

			for _, pool := range pools {
				cache.Add(pool.Hex())

				var (
					poolTokenRegistry common.Address
					priceQuoter       common.Address
				)
				err := chain.Provider.Client.CallCtx(
					ctx,
					eth.CallFunc(pool, tokenRegistryGetter).Returns(&poolTokenRegistry),
					eth.CallFunc(pool, quoterGetter).Returns(&priceQuoter),
				)
				if err != nil {
					return nil, err
				}
				cache.Add(priceQuoter.Hex())

				poolTokens, err := chain.GetAllTokensFromTokenIndex(ctx, poolTokenRegistry)
				if err != nil {
					return nil, err
				}

				for _, token := range poolTokens {
					cache.Add(token.Hex())
				}
				watchableIndex[poolTokenRegistry.Hex()] = true
			}

			watchableIndex[registryMap[celoutils.PoolIndex].Hex()] = true
		}
	}

	return watchableIndex, nil
}
