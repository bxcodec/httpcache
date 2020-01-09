package inmem_test

import (
	"testing"
	"time"

	"github.com/bxcodec/gotcha"
	inmemcache "github.com/bxcodec/gotcha/cache"
	"github.com/bxcodec/httpcache/cache"
	"github.com/bxcodec/httpcache/cache/inmem"
)

func TestCacheInMemory(t *testing.T) {
	c := gotcha.New(
		gotcha.NewOption().SetAlgorithm(inmemcache.LRUAlgorithm).
			SetExpiryTime(0).SetMaxSizeItem(100),
	)

	cacheObj := inmem.NewCache(c)
	testKey := "KEY"
	testVal := cache.CachedResponse{
		DumpedResponse: nil,
		RequestURI:     "http://bxcodec.io",
		RequestMethod:  "GET",
		CachedTime:     time.Now(),
	}

	// Try to SET item
	err := cacheObj.Set(testKey, testVal, time.Second*5)
	if err != nil {
		t.Fatalf("expected %v, got %v", nil, err)
	}

	// try to GET item from cache
	res, err := cacheObj.Get(testKey)
	if err != nil {
		t.Fatalf("expected %v, got %v", nil, err)
	}
	// assert the content
	if res.RequestURI != testVal.RequestURI {
		t.Fatalf("expected %v, got %v", testVal.RequestURI, res.RequestURI)
	}
	// assert the content
	if res.RequestMethod != testVal.RequestMethod {
		t.Fatalf("expected %v, got %v", testVal.RequestMethod, res.RequestMethod)
	}

	// try to DELETE the item
	err = cacheObj.Delete(testKey)
	if err != nil {
		t.Fatalf("expected %v, got %v", nil, err)
	}

	// try to re-GET item from cache after deleted
	res, err = cacheObj.Get(testKey)
	if err == nil {
		t.Fatalf("expected %v, got %v", err, nil)
	}
}
