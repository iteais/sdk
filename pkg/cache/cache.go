package cache

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/iteais/sdk/pkg"
	"github.com/redis/go-redis/v9"
)

func GetOrSet[T any](key string, f func() T, duration time.Duration) *T {
	val := *new(T)

	port, err := strconv.Atoi(os.Getenv("REDIS_DB"))

	if err != nil {
		pkg.App.Log.Warn("error parsing redis db", err)
		val = f()
		return &val
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "", // no password set
		DB:       port,
	})

	appCache := cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	ctx := context.TODO()
	err = appCache.Get(ctx, key, &val)

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
