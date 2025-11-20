package test

import (
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/tkahng/playground/internal/tools/utils"
)

type RandomSelector[T any] interface {
	Select() T
}

type RandomeSelectorImpl[T any] struct {
	options []T
	r       *rand.Rand
}

func (r *RandomeSelectorImpl[T]) Select() T {
	return r.options[r.r.Intn(len(r.options))]
}

func NewRandomeSelector[T any](options ...T) RandomSelector[T] {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return &RandomeSelectorImpl[T]{
		options: options,
		r:       r,
	}
}

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

func TestSliceItemsOrderByFunc[T any](t testing.TB, got []T, fn func(first T, second T) bool) {
	for i := 1; i < len(got)-1; i++ {
		if fn(got[i], got[i+1]) {
			continue
		}
		t.Errorf("test %s: items are not in order. first item %v > second item %v", t.Name(), got[i], got[i+1])
	}
}

func TestSliceEveryFunc[T any](t testing.TB, msg string, items []T, predicate func(item T) bool) {
	t.Helper()
	for idx, item := range items {
		if predicate(item) {
			continue
		}
		t.Errorf("test %s: %s - predicate failed for item %d: item %v", t.Name(), msg, idx, item)
	}
}

func TestSliceSomeFunc[T any](t *testing.T, msg string, items []T, predicate func(item T) bool) {
	t.Helper()
	if slices.ContainsFunc(items, predicate) {
		return
	}
	t.Errorf("test %s: %s - no items matched predicate", t.Name(), msg)
}
func TestSliceEveryUniqueFunc[T any, K comparable](t *testing.T, msg string, items []T, getKey func(T) K) {
	seen := make(map[K]struct{})
	for idx, item := range items {
		key := getKey(item)
		if _, exists := seen[key]; exists {
			t.Errorf("test %s: %s - item at index %d, key %v is not unique", t.Name(), msg, idx, key)
		}
		seen[key] = struct{}{}
	}
}
func ReturnResultOrFail[T any](t *testing.T, fn func() (T, error)) T {
	result, err := fn()
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// CompareFields compares specific fields of two structs (a and b) by name.
// It returns true if all specified fields are equal, along with an empty string.
// If any specified field is not equal, it returns false and a message detailing the first difference.
func CompareFields(a, b any, fieldNames ...string) (bool, string) {
	if len(fieldNames) == 0 {
		return reflect.DeepEqual(a, b), ""
	}
	// Ensure both inputs are structs and of the same concrete type
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	// Check if the inputs are pointers; if so, dereference them
	if valA.Kind() == reflect.Ptr {
		valA = valA.Elem()
	}
	if valB.Kind() == reflect.Ptr {
		valB = valB.Elem()
	}

	if valA.Kind() != reflect.Struct || valB.Kind() != reflect.Struct {
		return false, "Both inputs must be structs or pointers to structs"
	}

	if valA.Type() != valB.Type() {
		return false, fmt.Sprintf("Struct types are different: %v vs %v", valA.Type(), valB.Type())
	}

	for _, fieldName := range fieldNames {
		// Get the field by name
		fieldA := valA.FieldByName(fieldName)
		fieldB := valB.FieldByName(fieldName)

		// Check if the field exists
		if !fieldA.IsValid() || !fieldB.IsValid() {
			return false, fmt.Sprintf("Field '%s' not found in struct", fieldName)
		}

		// Compare the field values using DeepEqual
		// For slices ([]byte in your Log example) or complex types,
		// DeepEqual is the correct comparison method.
		if !reflect.DeepEqual(fieldA.Interface(), fieldB.Interface()) {
			return false, fmt.Sprintf(
				"Field '%s' differs: a is '%v', b is '%v'",
				fieldName,
				fieldA.Interface(),
				fieldB.Interface(),
			)
		}
	}

	return true, ""
}

func MustUnMarshal[T any](t testing.TB, data []byte) T {
	t.Helper()
	res, err := utils.UnmarshalJSON[T](data)
	if err != nil {
		t.Fatal(err)
	}
	return res
}
