package cache

import (
	"log/slog"

	"github.com/grassrootseconomics/celo-tracker/internal/chain"
)

type (
	Cache interface {
		Purge() error
		Exists(string) bool
		Add(string) bool
		Size() int
	}

	CacheOpts struct {
		Logg      *slog.Logger
		Chain     *chain.Chain
		CacheType string
	}
)

func New(o CacheOpts) Cache {
	var (
		cache Cache
	)

	switch o.CacheType {
	case "map":
		cache = NewMapCache()
	default:
		cache = NewMapCache()
	}
	o.Logg.Debug("bootstrapping cache")

	return cache
}
