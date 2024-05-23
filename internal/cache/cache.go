package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/grassrootseconomics/celo-tracker/internal/chain"
)

type (
	Cache interface {
		Purge() error
		Exists(string) bool
		Add(string, bool)
		Remove(string)
		IsWatchableIndex(string) bool
		Size() int
	}
	CacheOpts struct {
		Chain      chain.Chain
		Logg       *slog.Logger
		CacheType  string
		Blacklist  []string
		Registries []string
		Watchlist  []string
	}
)

func New(o CacheOpts) (Cache, error) {
	var cache Cache

	switch o.CacheType {
	case "map":
		cache = NewMapCache()
	default:
		cache = NewMapCache()
		o.Logg.Warn("invalid cache type, using default type (map)")
	}

	geSmartContracts, err := o.Chain.Provider().GetGESmartContracts(
		context.Background(),
		o.Registries,
	)
	if err != nil {
		return nil, fmt.Errorf("cache could not bootstrap GE smart contracts: err %v", err)
	}

	for k, v := range geSmartContracts {
		cache.Add(k, v)
	}
	for _, address := range o.Watchlist {
		cache.Add(address, false)
	}
	for _, address := range o.Blacklist {
		cache.Remove(address)
	}
	o.Logg.Info("cache bootstrap complete", "cached_addresses", cache.Size())

	return cache, nil
}
