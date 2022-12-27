package httpcache

import (
	"net/http"
	"time"

	"github.com/bxcodec/gotcha"
	inmemcache "github.com/bxcodec/gotcha/cache"
	"github.com/bxcodec/httpcache/cache"
	"github.com/bxcodec/httpcache/cache/inmem"
	rediscache "github.com/bxcodec/httpcache/cache/redis"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

// NewWithCustomStorageCache will initiate the httpcache with your defined cache storage
// To use your own cache storage handler, you need to implement the cache.Interactor interface
// And pass it to httpcache.
func NewWithCustomStorageCache(client *http.Client, rfcCompliance bool,
	cacheInteractor cache.ICacheInteractor) (cacheHandler *CacheHandler, err error) {
	return newClient(client, rfcCompliance, cacheInteractor)
}

func newClient(client *http.Client, rfcCompliance bool,
	cacheInteractor cache.ICacheInteractor) (cachedHandler *CacheHandler, err error) {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	cachedHandler = NewCacheHandlerRoundtrip(client.Transport, rfcCompliance, cacheInteractor)
	client.Transport = cachedHandler
	return
}

const (
	MaxSizeCacheItem = 100
)

// NewWithInmemoryCache will create a complete cache-support of HTTP client with using inmemory cache.
// If the duration not set, the cache will use LFU algorithm
func NewWithInmemoryCache(client *http.Client, rfcCompliance bool, duration ...time.Duration) (cachedHandler *CacheHandler, err error) {
	var expiryTime time.Duration
	if len(duration) > 0 {
		expiryTime = duration[0]
	}
	c := gotcha.New(
		gotcha.NewOption().SetAlgorithm(inmemcache.LRUAlgorithm).
			SetExpiryTime(expiryTime).SetMaxSizeItem(MaxSizeCacheItem),
	)

	return newClient(client, rfcCompliance, inmem.NewCache(c))
}

// NewWithRedisCache will create a complete cache-support of HTTP client with using redis cache.
// If the duration not set, the cache will use LFU algorithm
func NewWithRedisCache(client *http.Client, rfcCompliance bool, options *rediscache.CacheOptions,
	duration ...time.Duration) (cachedHandler *CacheHandler, err error) {
	var ctx = context.Background()
	var expiryTime time.Duration
	if len(duration) > 0 {
		expiryTime = duration[0]
	}
	c := redis.NewClient(&redis.Options{
		Addr:     options.Addr,
		Password: options.Password,
		DB:       options.DB,
	})

	return newClient(client, rfcCompliance, rediscache.NewCache(ctx, c, expiryTime))
}
