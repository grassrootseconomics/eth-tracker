package cache

import (
	"log/slog"

	"github.com/puzpuzpuz/xsync/v3"
)

type (
	MapCache struct {
		mapCache       *xsync.Map
		logg           *slog.Logger
		watchableIndex WatchableIndex
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

func (c *MapCache) Add(key string) {
	c.mapCache.Store(key, nil)
}

func (c *MapCache) Remove(key string) {
	c.mapCache.Delete(key)
}

func (c *MapCache) Size() int {
	return c.mapCache.Size()
}

func (c *MapCache) SetWatchableIndex(watchableIndex WatchableIndex) {
	c.watchableIndex = watchableIndex
}

func (c *MapCache) IsWatchableIndex(key string) bool {
	_, ok := c.watchableIndex[key]
	return ok
}
