package mystorage

import (
	"time"

	"github.com/bxcodec/httpcache/cache"
	patrickCache "github.com/patrickmn/go-cache"
)

type customInMemStorage struct {
	cacheHandler *patrickCache.Cache
}

// NewCustomInMemStorage will return a custom in memory cache
func NewCustomInMemStorage() cache.ICacheInteractor {
	return &customInMemStorage{
		cacheHandler: patrickCache.New(patrickCache.DefaultExpiration, time.Second*10),
	}
}

func (c customInMemStorage) Set(key string, value cache.CachedResponse) error {
	c.cacheHandler.Set(key, value, patrickCache.DefaultExpiration)
	return nil
}

func (c customInMemStorage) Get(key string) (res cache.CachedResponse, err error) {
	cachedRes, ok := c.cacheHandler.Get(key)
	if !ok {
		err = cache.ErrCacheMissed
		return
	}
	res, ok = cachedRes.(cache.CachedResponse)
	if !ok {
		err = cache.ErrInvalidCachedResponse
		return
	}
	return
}
func (c customInMemStorage) Delete(key string) error {
	c.cacheHandler.Delete(key)
	return nil
}
func (c customInMemStorage) Flush() error {
	c.cacheHandler.Flush()
	return nil
}
func (c customInMemStorage) Origin() string {
	return "MY-OWN-CUSTOM-INMEMORY-CACHED"
}
