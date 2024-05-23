package cache

import (
	"github.com/puzpuzpuz/xsync/v3"
)

type mapCache struct {
	xmap *xsync.Map
}

func NewMapCache() Cache {
	return &mapCache{
		xmap: xsync.NewMap(),
	}
}

func (c *mapCache) Purge() error {
	c.xmap.Clear()
	return nil
}

func (c *mapCache) Exists(key string) bool {
	_, ok := c.xmap.Load(key)
	return ok
}

func (c *mapCache) Add(key string, value bool) {
	c.xmap.Store(key, value)
}

func (c *mapCache) Remove(key string) {
	c.xmap.Delete(key)
}

func (c *mapCache) Size() int {
	return c.xmap.Size()
}

func (c *mapCache) IsWatchableIndex(key string) bool {
	watchable, ok := c.xmap.Load(key)
	if !ok {
		return false
	}
	watchableBool, ok := watchable.(bool)
	if !ok {
	}
	return watchableBool
}
