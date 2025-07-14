package cache

import (
	"context"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"os"
	"strings"
	"time"
)

func GetOrSet[T any](key string, f func() T) *T {

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

	mycache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	ctx := context.TODO()
	val := *new(T)
	err := mycache.Get(ctx, key, val)

	if err != nil {
		val = f()
		if err := mycache.Set(&cache.Item{
			Ctx:   ctx,
			Key:   key,
			Value: val,
			TTL:   time.Hour,
		}); err != nil {
			panic(err)
		}
	}
	return &val
}
