package memo

import (
	"context"
)

type MemoizedFunc[T any, K comparable] func(ctx context.Context, key K) (T, error)

type MemoizedStore[T any, K comparable] struct {
	memoizedFunc MemoizedFunc[T, K]
	cache        map[K]T
}

func NewMemoizedStore[T any, K comparable](memoizedFunc MemoizedFunc[T, K]) *MemoizedStore[T, K] {
	return &MemoizedStore[T, K]{
		memoizedFunc: memoizedFunc,
		cache:        make(map[K]T),
	}
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
