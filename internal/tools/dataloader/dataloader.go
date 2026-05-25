package dataloader

import (
	"context"
	"errors"
	"sync"
)

// ErrClosed is returned by Load when the dataloader has been closed.
var ErrClosed = errors.New("dataloader: closed")

type LoaderFunc[T any, K comparable] func(ctx context.Context, keys []K) ([]T, error)

type keyResult[T any, K comparable] struct {
	key        K
	resultChan chan result[T]
}

type result[T any] struct {
	value T
	err   error
}

// Dataloader batches concurrent Load calls into a single loader invocation.
//
// Lifecycle:
//  1. Goroutines call Load concurrently; each blocks until its result arrives.
//  2. Caller calls Close once all Load calls have been issued.
//  3. Caller calls Wait to process the batch and unblock all Load callers.
type Dataloader[T any, K comparable] struct {
	mu      sync.RWMutex
	keyChan chan keyResult[T, K]
	fn      LoaderFunc[T, K]
	once    sync.Once
	closed  bool
	ready   chan struct{} // closed by Close to unblock Wait
}

func New[T any, K comparable](fn LoaderFunc[T, K]) *Dataloader[T, K] {
	return &Dataloader[T, K]{
		keyChan: make(chan keyResult[T, K], 10000),
		fn:      fn,
		ready:   make(chan struct{}),
	}
}

// Load enqueues key for the next batch and blocks until Wait delivers the result.
// Returns ErrClosed if Close has already been called.
func (d *Dataloader[T, K]) Load(ctx context.Context, key K) (T, error) {
	resultChan := make(chan result[T], 1)

	// Hold RLock while sending so that Close's WLock waits for the send to
	// complete before marking the dataloader closed.
	d.mu.RLock()
	if d.closed {
		d.mu.RUnlock()
		var zero T
		return zero, ErrClosed
	}
	d.keyChan <- keyResult[T, K]{key: key, resultChan: resultChan}
	d.mu.RUnlock()

	select {
	case r := <-resultChan:
		return r.value, r.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// Close signals that no more keys will be submitted. Must be called exactly
// once, after all Load calls have been issued.
func (d *Dataloader[T, K]) Close() {
	d.once.Do(func() {
		// WLock waits for all in-progress Load sends to complete, then sets
		// closed so that subsequent Load calls return ErrClosed immediately.
		d.mu.Lock()
		d.closed = true
		d.mu.Unlock()
		close(d.ready)
	})
}

// Wait blocks until Close is called, then processes all enqueued keys in a
// single batch and delivers results to all blocked Load callers.
func (d *Dataloader[T, K]) Wait(ctx context.Context) error {
	select {
	case <-d.ready:
	case <-ctx.Done():
		return ctx.Err()
	}

	// By the time d.ready is closed, Close() has released the WLock with
	// closed=true. Every Load that held an RLock before that point has already
	// sent its key. No further sends can occur. Drain non-blockingly.
	var resKeys []keyResult[T, K]
	var keys []K
	for draining := true; draining; {
		select {
		case resK := <-d.keyChan:
			resKeys = append(resKeys, resK)
			keys = append(keys, resK.key)
		default:
			draining = false
		}
	}

	if len(keys) == 0 {
		return nil
	}

	results, err := d.fn(ctx, keys)

	loaderErr := err
	if err == nil && len(results) != len(keys) {
		loaderErr = errors.New("dataloader: loader returned wrong number of results")
	}

	for i, resKey := range resKeys {
		if loaderErr != nil {
			resKey.resultChan <- result[T]{err: loaderErr}
		} else {
			resKey.resultChan <- result[T]{value: results[i]}
		}
	}

	return loaderErr
}
