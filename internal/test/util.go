package test

func GetFirstItem[T any](items []T) T {
	if len(items) > 0 {
		return items[0]
	}
	var zero T
	return zero
}
