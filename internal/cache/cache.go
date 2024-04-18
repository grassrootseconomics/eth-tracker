package cache

import (
	"context"
	"log/slog"

	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/w3-celo"
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

	if err := bootstrapAllGESmartContracts(
		context.Background(),
		o.Registries,
		o.Chain,
		cache,
	); err != nil {
		return nil, err
	}
	o.Logg.Debug("cache bootstrap complete", "cached_addresses", cache.Size())

	return cache, nil
}
