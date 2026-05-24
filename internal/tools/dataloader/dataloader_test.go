//go:build !integration

package dataloader

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDataloader_Load(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	keys := []int{1, 2, 3}
	d := New(func(ctx context.Context, keys []int) ([]int, error) {
		return keys, nil
	})

	var mu sync.Mutex
	var newKeys []int
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
			newKeys = append(newKeys, res)
			mu.Unlock()
		}()
	}

	// spin until all keys are buffered in keyChan, then close to signal Wait
	for len(d.keyChan) < len(keys) {
		time.Sleep(time.Millisecond)
	}
	close(d.keyChan)

	err := d.Wait(ctx)
	assert.NoError(t, err)
	wg.Wait()
	assert.Len(t, newKeys, len(keys))
}
