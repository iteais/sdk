package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"time"
)

var ctx context.Context

func init() {
	ctx = context.Background()
}

func GetOrSet[T any](key string, f func() T) *T {
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	gocacheStore := go_cache.NewGoCache(gocacheClient)

	cacheManager := cache.New[[]byte](gocacheStore)

	cachedValue, err := cacheManager.Get(ctx, key)
	if err == nil {

		decoded := new(T)

		buffer := bytes.NewBuffer(cachedValue)
		dec := gob.NewDecoder(buffer)
		err = dec.Decode(&decoded)
		if err != nil {
			panic(err)
		}

		return decoded
	}

	value := f()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(value)
	if err != nil {
		panic(err)
	}

	err = cacheManager.Set(ctx, key, buf.Bytes())
	if err != nil {
		panic(err)
	}

	return &value
}
