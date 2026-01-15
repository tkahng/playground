package memo

import (
	"context"
)

// GetFunc is a function that fetches a resource by id.
type GetFunc[T any, K comparable] func(ctx context.Context, key K) (T, error)

// LoadFunc is a function that fetches multiple resources by ids.
type LoadFunc[T any, K comparable] func(ctx context.Context, keys ...K) ([]T, error)

// MemoizedStore is a key value store that has methods to get a value by key and load values by keys.
// When called, MemoizedStore.Get will first look at the cache, return it if found, else will call the MemoizedStore.memoizedFunc,
// cache that result so subsequent calls to the same key wouldn't have to waste resources.
// MemoizedStore.Load will fetch all resources of the supplied keys, then cache them.
type MemoizedStore[T any, K comparable] struct {
	keyFunc      func(T) K
	loadFunc     LoadFunc[T, K]
	memoizedFunc GetFunc[T, K]
	cache        map[K]T
}

func New[T any, K comparable](memoizedFunc GetFunc[T, K], loadFunc LoadFunc[T, K], keyFunc func(T) K) *MemoizedStore[T, K] {
	return &MemoizedStore[T, K]{
		keyFunc:      keyFunc,
		loadFunc:     loadFunc,
		memoizedFunc: memoizedFunc,
		cache:        make(map[K]T),
	}
}

func (ms *MemoizedStore[T, K]) Load(ctx context.Context, keys ...K) error {
	res, err := ms.loadFunc(ctx, keys...)
	if err != nil {
		return err
	}

	for _, r := range res {
		k := ms.keyFunc(r)
		ms.cache[k] = r
	}

	return nil
}
func (ms *MemoizedStore[T, K]) Get(ctx context.Context, key K) (T, error) {
	if val, ok := ms.cache[key]; ok {
		return val, nil
	}

	val, err := ms.memoizedFunc(ctx, key)
	if err != nil {
		var zero T
		return zero, err
	}

	ms.cache[key] = val
	return val, nil
}
