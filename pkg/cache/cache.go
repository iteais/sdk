package cache

import (
	"context"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"time"
)

func GetOrSet[T any](key string, f func() T) T {
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	gocacheStore := go_cache.NewGoCache(gocacheClient)

	cacheManager := cache.New[T](gocacheStore)

	cachedValue, err := cacheManager.Get(context.Background(), key)
	if err == nil {
		return cachedValue
	}

	value := f()

	err = cacheManager.Set(context.Background(), key, value)
	if err != nil {
		panic(err)
	}

	return value
}
