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
)

// Interactor ...
type Interactor interface {
	Set(key string, value CachedResponse) error
	Get(key string) (CachedResponse, error)
	Delete(key string) error
}

// CachedResponse represent the cacher struct item
type CachedResponse struct {
	StatusCode     int       `json:"statusCode"`
	DumpedResponse []byte    `json:"response"`
	DumpedBody     []byte    `json:"body"`
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
