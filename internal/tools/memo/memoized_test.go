package memo

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testGetterLoader struct {
	calls map[string]int
	loads map[string]int
}

func (tg *testGetterLoader) Get(ctx context.Context, key string) (string, error) {
	tg.calls[key]++
	return "value for " + key, nil
}
func (tg *testGetterLoader) Load(ctx context.Context, keys ...string) ([]string, error) {
	uniqueKeys := make([]string, len(keys))
	for _, key := range keys {
		// For Load, we'll increment a separate counter to distinguish from Get calls
		uniqueKeys = append(uniqueKeys, "value for "+key)
		tg.loads[key]++
	}
	return uniqueKeys, nil
}

func TestMemoizedStore_Get(t *testing.T) {
	ctx := context.Background()
	getter := &testGetterLoader{calls: make(map[string]int), loads: make(map[string]int)}
	ms := New(getter.Get, getter.Load, func(key string) string { return strings.ReplaceAll(key, "value for ", "") })

	keys := []string{"a", "a", "b", "b", "c"}
	for _, key := range keys {
		val, err := ms.Get(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedVal := "value for " + key
		if val != expectedVal {
			t.Errorf("expected %q, got %q", expectedVal, val)
		}
	}

	for key, count := range getter.calls {
		if count != 1 {
			t.Errorf("expected 1 call for key %q, got %d", key, count)
		}
	}
}

func TestMemoizedStore_Get_with_load(t *testing.T) {
	ctx := context.Background()
	getter := &testGetterLoader{calls: make(map[string]int), loads: make(map[string]int)}
	ms := New(getter.Get, getter.Load, func(key string) string { return strings.ReplaceAll(key, "value for ", "") })

	keys := []string{"a", "a", "b", "b", "c"}
	err := ms.Load(ctx, keys...)
	assert.NoError(t, err)
	for _, key := range keys {
		val, err := ms.Get(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedVal := "value for " + key
		if val != expectedVal {
			t.Errorf("expected %q, got %q", expectedVal, val)
		}
	}

	for key, count := range getter.calls {
		if count != 0 {
			t.Errorf("expected 1 call for key %q, got %d", key, count)
		}
	}
	assert.Equal(t, len(getter.loads), 3)
}
