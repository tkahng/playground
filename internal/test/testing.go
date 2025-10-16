package test

import "testing"

func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long running test in short mode")
	}
}

func GetFirstItem[T any](t *testing.T, items []T) T {
	t.Helper()
	if len(items) <= 0 {
		t.Errorf("items lengths is zero.")
	}
	return items[0]
}

func TestSliceLength[M any](t *testing.T, got []*M, expected int) {
	t.Helper()
	if len(got) != expected {
		t.Errorf("%s: check slice length got = %d, want %d", t.Name(), len(got), expected)
	}
}

func TestSliceItemsOrderByFunc[T any](t *testing.T, got []T, fn func(first T, second T) bool) {
	for i := 1; i < len(got)-1; i++ {
		// firstName, secondName := *got[i].Name, *got[i+1].Name
		// if firstName > secondName {
		// 	t.Errorf("users are not in order. first name %s > second name %s", firstName, secondName)
		// }
		if fn(got[i], got[i+1]) {
			continue
		}
		t.Errorf("test %s: items are not in order. first item %v > second item %v", t.Name(), got[i], got[i+1])
	}
}

func TestSliceItemsByFunc[T any](t *testing.T, predicateName string, items []T, predicate func(item T) bool) {
	t.Helper()
	for _, item := range items {
		if predicate(item) {
			continue
		}
		t.Errorf("test %s: predicate %s returned false for item %v", t.Name(), predicateName, item)
	}
}
