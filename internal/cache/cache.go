package cache

import (
	"context"
	"log/slog"

	"github.com/grassrootseconomics/celo-tracker/pkg/chain"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	Cache interface {
		Purge() error
		Exists(string) bool
		Add(string)
		Remove(string)
		SetWatchableIndex(WatchableIndex)
		ISWatchAbleIndex(string) bool
		Size() int
	}

	CacheOpts struct {
		Logg       *slog.Logger
		Chain      *chain.Chain
		CacheType  string
		Registries []string
		Blacklist  []string
		Watchlist  []string
	}

	WatchableIndex map[string]bool
)

var (
	tokenRegistryGetter = w3.MustNewFunc("tokenRegistry()", "address")
	quoterGetter        = w3.MustNewFunc("quoter()", "address")
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

	watchableIndex, err := bootstrapGESmartContracts(
		context.Background(),
		o.Registries,
		o.Chain,
		cache,
	)
	if err != nil {
		return nil, err
	}
	// We only watch the token and pool indexes
	// If at some point we want to watch the user index, this line should be removed
	cache.SetWatchableIndex(watchableIndex)

	for _, address := range o.Watchlist {
		cache.Add(address)
	}
	for _, address := range o.Blacklist {
		cache.Remove(address)
	}
	o.Logg.Debug("cache bootstrap complete", "cached_addresses", cache.Size())

	return cache, nil
}
