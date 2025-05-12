package mapper

import (
	"reflect"
	"testing"
)

type testStruct struct {
	ID   int
	Name string
}

func TestMapTo(t *testing.T) {
	records := []testStruct{
		{ID: 1, Name: "one"},
		{ID: 2, Name: "two"},
		{ID: 3, Name: "three"},
	}
	keys := []int{1, 2, 4}

	result := MapTo(records, keys, func(r testStruct) int { return r.ID })

	expected := []*testStruct{
		{ID: 1, Name: "one"},
		{ID: 2, Name: "two"},
		nil,
	}

	for i := range expected {
		if expected[i] == nil && result[i] != nil {
			t.Errorf("Expected nil at index %d but got %v", i, result[i])
			continue
		}
		if expected[i] != nil && result[i] == nil {
			t.Errorf("Expected %v at index %d but got nil", expected[i], i)
			continue
		}
		if expected[i] != nil && !reflect.DeepEqual(*expected[i], *result[i]) {
			t.Errorf("Expected %v at index %d but got %v", *expected[i], i, *result[i])
		}
	}
}

func TestMap(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}

	result := Map(numbers, func(n int) int { return n * 2 })

	expected := []int{2, 4, 6, 8, 10}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

func TestFilter(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}

	result := Filter(numbers, func(n int) bool { return n%2 == 0 })

	expected := []int{2, 4}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

func TestReduce(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}

	result := Reduce(numbers, 0, func(acc, curr int) int { return acc + curr })

	expected := 15

	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}
