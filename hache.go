package hache

import (
	"errors"
	"net/http"
	"time"
)

var (
	// ErrInvalidCachedResponse will throw if the cached response is invalid
	ErrInvalidCachedResponse = errors.New("Cached Response is Invalid")
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

// CachedResponse represent the cacher struct item
type CachedResponse struct {
	StatusCode     int       `json:"statusCode"`
	DumpedResponse []byte    `json:"body"`
	RequestURI     string    `json:"requestUri"`
	RequestMethod  string    `json:"requestMethod"`
	CachedTime     time.Time `json:"cachedTime"`
}

// Validate will validate the cached response
func (c *CachedResponse) Validate() (err error) {
	if c.StatusCode == 0 {
		return ErrInvalidCachedResponse
	}

	if c.RequestMethod == "" {
		return ErrInvalidCachedResponse
	}

	if c.RequestURI == "" {
		return ErrInvalidCachedResponse
	}

	if len(c.DumpedResponse) == 0 {
		return ErrInvalidCachedResponse
	}

	if c.CachedTime.IsZero() {
		return ErrInvalidCachedResponse
	}

	return
}
