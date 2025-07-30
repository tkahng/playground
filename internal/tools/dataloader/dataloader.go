package dataloader

import "context"

type LoaderFunc[T any, K comparable] func(ctx context.Context, keys []K) ([]T, error)

type KeyResult[T any, K comparable] struct {
	Key        K
	ResultChan chan T
}

type Dataloader[T any, K comparable] struct {
	keyChan chan KeyResult[T, K]
	fn      LoaderFunc[T, K]
}

func New[T any, K comparable](fn LoaderFunc[T, K]) *Dataloader[T, K] {
	return &Dataloader[T, K]{
		keyChan: make(chan KeyResult[T, K], 10000),
		fn:      fn,
	}
}

func (d *Dataloader[T, K]) Load(ctx context.Context, key K) (T, error) {
	resultChan := make(chan T)
	d.keyChan <- KeyResult[T, K]{Key: key, ResultChan: resultChan}
	return <-resultChan, nil
}

func (d *Dataloader[T, K]) Wait(ctx context.Context) error {
	var resKeys []KeyResult[T, K]
	var keys []K
	for resK := range d.keyChan {
		resKeys = append(resKeys, resK)
		keys = append(keys, resK.Key)
	}

	results, err := d.fn(ctx, keys)
	if err != nil {
		return err
	}

	for i, resKey := range resKeys {
		resKey.ResultChan <- results[i]
	}
	return nil
}
