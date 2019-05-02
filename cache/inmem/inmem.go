package inmem

import (
	memcache "github.com/bxcodec/gotcha/cache"
	"github.com/bxcodec/hache/cache"
)

type inmemCache struct {
	c memcache.Cache
}

// NewCache ...
func NewCache(c memcache.Cache) cache.Interactor {
	return &inmemCache{
		c: c,
	}
}

func (i *inmemCache) Set(key string, value cache.CachedResponse) (err error) {
	panic("TODO")
	return
}

func (i *inmemCache) Get(key string) (res cache.CachedResponse, err error) {
	panic("TODO")
	return
}

func (i *inmemCache) Delete(key string) (err error) {
	panic("TODO")
	return
}
