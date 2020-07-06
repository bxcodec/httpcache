package cache

import (
	"errors"
	"time"
)

var (
	// ErrInvalidCachedResponse will throw if the cached response is invalid
	ErrInvalidCachedResponse = errors.New("Cached Response is Invalid")
	// ErrFailedToSaveToCache will throw if the item can't be saved to cache
	ErrFailedToSaveToCache = errors.New("Failed to save item")
	// ErrCacheMissed will throw if an item can't be retrieved (due to invalid, or missing)
	ErrCacheMissed = errors.New("Cache is missing")
	// ErrStorageInternal will throw when some internal error in storage occurred
	ErrStorageInternal = errors.New("Internal error in storage")
)

// Cache storage type
const (
	CacheStorageInMemory = "IN-MEMORY"
	CacheRedis           = "REDIS"
	// TODO (bxcodec): Add another storage type
)

// ICacheInteractor ...
type ICacheInteractor interface {
	Set(key string, value CachedResponse) error
	Get(key string) (CachedResponse, error)
	Delete(key string) error
	Flush() error
	Origin() string
}

// CachedResponse represent the cacher struct item
type CachedResponse struct {
	DumpedResponse []byte    `json:"response"`      // The dumped response body
	RequestURI     string    `json:"requestUri"`    // The requestURI of the response
	RequestMethod  string    `json:"requestMethod"` // The HTTP Method that call the request for this response
	CachedTime     time.Time `json:"cachedTime"`    // The timestamp when this response is Cached
}

// Validate will validate the cached response
func (c *CachedResponse) Validate() (err error) {
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
