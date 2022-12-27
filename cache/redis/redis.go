package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bxcodec/httpcache/cache"
	"github.com/go-redis/redis/v8"
)

// CacheOptions for storing data for Redis connections
type CacheOptions struct {
	Addr     string
	Password string
	DB       int // 0 for default DB
}

type redisCache struct {
	ctx        context.Context
	cache      *redis.Client
	expiryTime time.Duration
}

// NewCache will return the redis cache handler
func NewCache(ctx context.Context, c *redis.Client, exptime time.Duration) cache.ICacheInteractor {
	return &redisCache{
		ctx:        ctx,
		cache:      c,
		expiryTime: exptime,
	}
}

func (i *redisCache) Set(key string, value cache.CachedResponse) (err error) { //nolint
	valueJSON, _ := json.Marshal(value)
	set := i.cache.Set(i.ctx, key, string(valueJSON), i.expiryTime*time.Second)
	if err := set.Err(); err != nil {
		fmt.Println(err)
		return cache.ErrStorageInternal
	}
	return nil
}

func (i *redisCache) Get(key string) (res cache.CachedResponse, err error) {
	get := i.cache.Do(i.ctx, "get", key)
	if err = get.Err(); err != nil {
		if err == redis.Nil {
			return cache.CachedResponse{}, cache.ErrCacheMissed
		}
		return cache.CachedResponse{}, cache.ErrStorageInternal
	}
	val := get.Val().(string)
	err = json.Unmarshal([]byte(val), &res)
	if err != nil {
		return cache.CachedResponse{}, cache.ErrStorageInternal
	}
	return
}

func (i *redisCache) Delete(key string) (err error) {
	// deleting in redis equal to setting expiration time for key to 0
	set := i.cache.Set(i.ctx, key, nil, 0)
	if err := set.Err(); err != nil {
		return cache.ErrStorageInternal
	}
	return nil
}

func (i *redisCache) Origin() string {
	return cache.CacheRedis
}

func (i *redisCache) Flush() error {
	flush := i.cache.FlushAll(i.ctx)
	if err := flush.Err(); err != nil {
		return cache.ErrStorageInternal
	}
	return nil
}
