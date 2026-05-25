package dataloader

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// waitForQueue spins until the dataloader's internal queue holds at least n
// items. Because sends from Load() complete while holding an RLock, once n
// items appear in the buffer we know those goroutines have released their
// RLock and are now blocked on resultChan, making it safe to call Close().
func waitForQueue[T any, K comparable](d *Dataloader[T, K], n int) {
	for len(d.keyChan) < n {
		runtime.Gosched()
	}
}

func TestDataloader_Load(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	keys := []int{1, 2, 3}
	d := New(func(ctx context.Context, keys []int) ([]int, error) {
		return keys, nil
	})

	var mu sync.Mutex
	var results []int
	var wg sync.WaitGroup

	for _, key := range keys {
		wg.Add(1)
		key := key
		go func() {
			defer wg.Done()
			res, err := d.Load(ctx, key)
			if err != nil {
				t.Error(err)
				return
			}
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}()
	}

	waitForQueue(d, len(keys))
	d.Close()

	err := d.Wait(ctx)
	assert.NoError(t, err)
	wg.Wait()
	assert.Len(t, results, len(keys))
}

// TestDataloader_ErrorPropagated verifies that a loader error is returned to
// every blocked Load caller rather than leaking goroutines.
func TestDataloader_ErrorPropagated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	loaderErr := errors.New("loader failed")
	d := New(func(ctx context.Context, keys []int) ([]int, error) {
		return nil, loaderErr
	})

	var wg sync.WaitGroup
	errs := make([]error, 3)
	for i := range errs {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			_, err := d.Load(ctx, i+1)
			errs[i] = err
		}()
	}

	waitForQueue(d, len(errs))
	d.Close()

	err := d.Wait(ctx)
	assert.ErrorIs(t, err, loaderErr)
	wg.Wait()
	for i, err := range errs {
		assert.ErrorIs(t, err, loaderErr, "goroutine %d: Load() should propagate loader error", i)
	}
}

// TestDataloader_WrongResultCount verifies that a mismatched result count is
// treated as an error and propagated to all Load callers.
func TestDataloader_WrongResultCount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	d := New(func(ctx context.Context, keys []int) ([]int, error) {
		return []int{keys[0]}, nil // intentionally returns fewer results
	})

	var wg sync.WaitGroup
	errs := make([]error, 3)
	for i := range errs {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			_, err := d.Load(ctx, i+1)
			errs[i] = err
		}()
	}

	waitForQueue(d, len(errs))
	d.Close()

	err := d.Wait(ctx)
	require.Error(t, err)
	wg.Wait()
	for i, err := range errs {
		assert.Error(t, err, "goroutine %d: Load() should propagate result-count mismatch", i)
	}
}

// TestDataloader_EmptyBatch verifies that Wait with no enqueued keys is a no-op.
func TestDataloader_EmptyBatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	called := false
	d := New(func(ctx context.Context, keys []int) ([]int, error) {
		called = true
		return nil, nil
	})
	d.Close()
	err := d.Wait(ctx)
	assert.NoError(t, err)
	assert.False(t, called, "loader should not be called for empty batch")
}

// TestDataloader_LoadAfterClose verifies that Load returns ErrClosed when
// called after Close.
func TestDataloader_LoadAfterClose(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	d := New(func(ctx context.Context, keys []int) ([]int, error) {
		return keys, nil
	})
	d.Close()
	d.Wait(ctx) //nolint:errcheck
	_, err := d.Load(ctx, 1)
	assert.ErrorIs(t, err, ErrClosed)
}
