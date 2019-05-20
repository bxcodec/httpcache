package hache

import (
	"net/http"
	"time"

	"github.com/bxcodec/gotcha"
	inmemcache "github.com/bxcodec/gotcha/cache"
	"github.com/bxcodec/hache/cache"
	"github.com/bxcodec/hache/cache/inmem"
)

// New ...
func New(client *http.Client, cacheInteractor cache.Interactor) (err error) {
	newClient(client, cacheInteractor)
	return
}

func newClient(client *http.Client, cacheInteractor cache.Interactor) (err error) {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	client.Transport = NewRoundtrip(client.Transport, cacheInteractor)
	return
}

// NewWithInmemoryCache will create a complete cache-support of HTTP client with using inmemory cache.
// If the duration not set, the cache will use LFU algorithm
func NewWithInmemoryCache(client *http.Client, duration ...time.Duration) (err error) {
	var expiryTime time.Duration
	if len(duration) > 0 {
		expiryTime = duration[0]
	}
	c := gotcha.New(
		gotcha.NewOption().SetAlgorithm(inmemcache.LRUAlgorithm).
			SetExpiryTime(expiryTime).SetMaxSizeItem(100),
	)

	return newClient(client, inmem.NewCache(c))
}
