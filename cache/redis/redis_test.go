package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/bxcodec/httpcache/cache"
	rediscache "github.com/bxcodec/httpcache/cache/redis"
	"github.com/redis/go-redis/v9"
)

func TestCacheRedis(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()
	c := redis.NewClient(&redis.Options{
		Addr:     s.Addr(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	cacheObj := rediscache.NewCache(context.Background(), c, 15)
	testKey := "KEY"
	testVal := cache.CachedResponse{
		DumpedResponse: nil,
		RequestURI:     "http://bxcodec.io",
		RequestMethod:  "GET",
		CachedTime:     time.Now(),
	}

	// Try to SET item
	err = cacheObj.Set(testKey, testVal)
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
