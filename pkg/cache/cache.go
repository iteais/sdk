package cache

import (
	"context"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"os"
	"strings"
	"time"
)

func GetOrSet[T any](key string, f func() T, duration time.Duration) *T {

	srv := os.Getenv("CACHE_SERVER")

	if srv == "" {
		val := f()
		return &val
	}

	parts := strings.Split(srv, ":")
	if len(parts) != 2 {
		val := f()
		return &val
	}

	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			parts[0]: ":" + parts[1],
		},
	})

	appCache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	ctx := context.TODO()
	val := *new(T)
	err := appCache.Get(ctx, key, &val)

	if err != nil {
		val = f()
		_ = appCache.Set(&cache.Item{
			Ctx:   ctx,
			Key:   key,
			Value: val,
			TTL:   duration,
		})
	}
	return &val
}
