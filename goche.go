package hache

import (
	"net/http"
	"time"
)

// New ...
func New(client *http.Client, cacheInteractor CacheInteractor) (err error) {
	roundtrip := &RoundTrip{
		DefaultRoundTripper: client.Transport,
		CacheInteractor:     cacheInteractor,
	}
	client.Transport = roundtrip
	return
}

// NewWithInmemoryCache will create a complete cache-support of HTTP client with using inmemory cache.
// If the duration not set, the cache will use LFU algorithm
func NewWithInmemoryCache(client *http.Client, duration ...time.Duration) (err error) {
	panic("TODO: (bxcodec)")
	return
}
