package cache

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celoutils/v3"
	"github.com/grassrootseconomics/eth-tracker/internal/chain"
	"github.com/grassrootseconomics/w3-celo"
	"github.com/grassrootseconomics/w3-celo/module/eth"
)

func bootstrapCache(
	chain chain.Chain,
	cache Cache,
	registries []string,
	watchlist []string,
	blacklist []string,
	lo *slog.Logger,
) error {
	var (
		tokenRegistryGetter = w3.MustNewFunc("tokenRegistry()", "address")
		quoterGetter        = w3.MustNewFunc("quoter()", "address")
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	for _, registry := range registries {
		registryMap, err := chain.Provider().RegistryMap(ctx, celoutils.HexToAddress(registry))
		if err != nil {
			lo.Error("could not fetch registry", "registry", registry, "error", err)
			os.Exit(1)
		}

		for k, v := range registryMap {
			if v != celoutils.ZeroAddress {
				if err := cache.Add(ctx, v.Hex()); err != nil {
					return err
				}

				lo.Debug("cached registry entry", "type", k, "address", v.Hex())
			}
		}

		if accountIndex := registryMap[celoutils.AccountIndex]; accountIndex != celoutils.ZeroAddress {
			if err := cache.Add(ctx, accountIndex.Hex()); err != nil {
				return err
			}

			lo.Debug("cached account index", "address", accountIndex.Hex())

			accountIndexIter, err := chain.Provider().NewBatchIterator(ctx, accountIndex)
			if err != nil {
				return err
			}
			for {
				accountIndexBatch, err := accountIndexIter.Next(ctx)
				if err != nil {
					return err
				}
				if accountIndexBatch == nil {
					break
				}

				for _, address := range accountIndexBatch {
					if address != celoutils.ZeroAddress {
						if err := cache.Add(ctx, address.Hex()); err != nil {
							return err
						}

					}
				}
				lo.Debug("cached account index batch", "batch_size", len(accountIndexBatch))
			}
		}

		if tokenIndex := registryMap[celoutils.TokenIndex]; tokenIndex != celoutils.ZeroAddress {
			if err := cache.Add(ctx, tokenIndex.Hex()); err != nil {
				return err
			}

			lo.Debug("cached token index", "address", tokenIndex.Hex())

			tokenIndexIter, err := chain.Provider().NewBatchIterator(ctx, tokenIndex)
			if err != nil {
				return err
			}
			for {
				tokenIndexBatch, err := tokenIndexIter.Next(ctx)
				if err != nil {
					return err
				}
				if tokenIndexBatch == nil {
					break
				}

				for _, address := range tokenIndexBatch {
					if address != celoutils.ZeroAddress {
						if err := cache.Add(ctx, address.Hex()); err != nil {
							return err
						}

					}
				}
				lo.Debug("cached token index batch", "batch_size", len(tokenIndexBatch))
			}
		}

		if poolIndex := registryMap[celoutils.PoolIndex]; poolIndex != celoutils.ZeroAddress {
			if err := cache.Add(ctx, poolIndex.Hex()); err != nil {
				return err
			}

			lo.Debug("cached pool index", "address", poolIndex.Hex())

			poolIndexIter, err := chain.Provider().NewBatchIterator(ctx, poolIndex)
			if err != nil {
				return err
			}
			for {
				poolIndexBatch, err := poolIndexIter.Next(ctx)
				if err != nil {
					return err
				}
				if poolIndexBatch == nil {
					break
				}
				for _, address := range poolIndexBatch {
					if address != celoutils.ZeroAddress {
						if err := cache.Add(ctx, address.Hex()); err != nil {
							return err
						}

						var poolTokenIndex, priceQuoter common.Address
						err := chain.Provider().Client.CallCtx(
							ctx,
							eth.CallFunc(address, tokenRegistryGetter).Returns(&poolTokenIndex),
							eth.CallFunc(address, quoterGetter).Returns(&priceQuoter),
						)
						if err != nil {
							return err
						}
						if priceQuoter != celoutils.ZeroAddress {
							if err := cache.Add(ctx, priceQuoter.Hex()); err != nil {
								return err
							}

							lo.Debug("cached pool index quoter", "pool", poolIndex.Hex(), "address", priceQuoter.Hex())
						}
						if poolTokenIndex != celoutils.ZeroAddress {
							if err := cache.Add(ctx, poolTokenIndex.Hex()); err != nil {
								return err
							}

							lo.Debug("cached pool index token index", "pool", poolIndex.Hex(), "address", poolTokenIndex.Hex())

							poolTokenIndexIter, err := chain.Provider().NewBatchIterator(ctx, poolTokenIndex)
							if err != nil {
								return err
							}
							for {
								poolTokenIndexBatch, err := poolTokenIndexIter.Next(ctx)
								if err != nil {
									return err
								}
								if poolTokenIndexBatch == nil {
									break
								}
								for _, address := range poolTokenIndexBatch {
									if address != celoutils.ZeroAddress {
										if err := cache.Add(ctx, address.Hex()); err != nil {
											return err
										}

									}
								}
								lo.Debug("cached pool token index batch", "batch_size", len(poolTokenIndexBatch))
							}
						}
					}
				}
				lo.Debug("cached pool index batch", "batch_size", len(poolIndexBatch))
			}
		}

		for _, address := range watchlist {
			if err := cache.Add(ctx, celoutils.HexToAddress(address).Hex()); err != nil {
				return err
			}
		}
		for _, address := range blacklist {
			if err := cache.Remove(ctx, celoutils.HexToAddress(address).Hex()); err != nil {
				return err
			}
		}
		if err := cache.Remove(ctx, celoutils.ZeroAddress.Hex()); err != nil {
			return err
		}
		cacheSize, err := cache.Size(ctx)
		if err != nil {
			return err
		}
		lo.Info("registry bootstrap complete", "registry", registry, "current_cache_size", cacheSize)
	}

	return nil
}
