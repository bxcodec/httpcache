package lru

import (
	"github.com/bxcodec/hache"
	lru "github.com/hashicorp/golang-lru"
)

var (
	// LruSize indicate the size of LRU item
	LruSize = 200
)

// NewCache implementa  inmem cache mecahnism with LRU Algorithm
func NewCache() (cache *Cache, err error) {
	lruCache, err := lru.New(LruSize)
	if err != nil {
		return
	}

	cache = &Cache{
		CacheSystem: lruCache,
	}
	return
}

// Cache ...
type Cache struct {
	CacheSystem *lru.Cache
}

// Set will save an item to cache system
func (c Cache) Set(key string, value interface{}) (err error) {
	ok := c.CacheSystem.Add(key, value)
	if !ok {
		err = hache.ErrFailedToSaveToCache
	}
	return
}

// Get will get an item from cache system
func (c Cache) Get(key string) (res interface{}, err error) {
	res, ok := c.CacheSystem.Get(key)
	if !ok {
		err = hache.ErrCacheMissed
	}
	return
}
