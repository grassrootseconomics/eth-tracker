package cache

import (
	"context"
	"log/slog"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celoutils/v2"
	"github.com/grassrootseconomics/w3-celo"
	"github.com/grassrootseconomics/w3-celo/module/eth"
)

type (
	Cache interface {
		Purge() error
		Exists(string) bool
		Add(string) bool
		Size() int
	}

	CacheOpts struct {
		Logg       *slog.Logger
		Chain      *chain.Chain
		CacheType  string
		Registries []string
	}
)

var (
	tokenRegistryFunc = w3.MustNewFunc("tokenRegistry()", "address")
)

func New(o CacheOpts) (Cache, error) {
	var (
		cache Cache
	)

	switch o.CacheType {
	case "map":
		cache = NewMapCache()
	default:
		cache = NewMapCache()
	}

	ctx := context.Background()
	for _, registry := range o.Registries {
		registryMap, err := o.Chain.Provider.RegistryMap(ctx, w3.A(registry))
		if err != nil {
			return nil, err
		}

		for _, v := range registryMap {
			cache.Add(v.Hex())
		}

		if registryMap[celoutils.TokenIndex] != common.ZeroAddress {
			tokens, err := o.Chain.GetAllTokensFromTokenIndex(ctx, registryMap[celoutils.TokenIndex])
			if err != nil {
				return nil, err
			}

			for _, token := range tokens {
				cache.Add(token.Hex())
			}
		}

		if registryMap[celoutils.PoolIndex] != common.ZeroAddress {
			pools, err := o.Chain.GetAllTokensFromTokenIndex(ctx, registryMap[celoutils.PoolIndex])
			if err != nil {
				return nil, err
			}

			for _, pool := range pools {
				cache.Add(pool.Hex())

				var (
					poolTokenRegistry common.Address
				)
				err := o.Chain.Provider.Client.CallCtx(
					ctx,
					eth.CallFunc(pool, tokenRegistryFunc).Returns(&poolTokenRegistry),
				)
				if err != nil {
					return nil, err
				}

				poolTokens, err := o.Chain.GetAllTokensFromTokenIndex(ctx, poolTokenRegistry)
				if err != nil {
					return nil, err
				}

				for _, token := range poolTokens {
					cache.Add(token.Hex())
				}
			}
		}
	}
	o.Logg.Debug("cache bootstrap complete", "cached_addresses", cache.Size())

	return cache, nil
}
