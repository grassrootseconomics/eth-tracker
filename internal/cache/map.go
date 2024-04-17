package cache

import (
	"log/slog"

	"github.com/puzpuzpuz/xsync/v3"
)

type (
	MapCache struct {
		mapCache *xsync.Map
		logg     *slog.Logger
	}
)

func NewMapCache() *MapCache {
	return &MapCache{
		mapCache: xsync.NewMap(),
	}
}

func (c *MapCache) Purge() error {
	c.mapCache.Clear()
	return nil
}

func (c *MapCache) Exists(key string) bool {
	_, ok := c.mapCache.Load(key)
	return ok
}

func (c *MapCache) Add(key string) bool {
	c.mapCache.Store(key, nil)
	return true
}

func (c *MapCache) Size() int {
	return c.mapCache.Size()
}
