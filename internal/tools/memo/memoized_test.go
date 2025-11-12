package memo

import (
	"context"
	"testing"
)

type testGetter struct {
	calls map[string]int
}

func (tg *testGetter) Get(ctx context.Context, key string) (string, error) {
	tg.calls[key]++
	return "value for " + key, nil
}

func TestMemoizedStore_Get(t *testing.T) {
	ctx := context.Background()
	getter := &testGetter{calls: make(map[string]int)}
	ms := NewMemoizedStore(getter.Get)

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
