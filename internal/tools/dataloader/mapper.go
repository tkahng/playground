package dataloader

// func MapTo[]

type KeyFn[T any, K comparable] func(T) K

// These helper functions are intended to be used in data loaders for mapping
// entity keys to entity values (db records). See `../context.ts`.
// https://github.com/graphql/dataloader

// export function mapTo<R, K>(
//   records: ReadonlyArray<R>,
//   keys: ReadonlyArray<K>,
//   keyFn: (record: R) => K,
// ): Array<R | null> {
//   const map = new Map(records.map((x) => [keyFn(x), x]));
//   return keys.map((key) => map.get(key) || null);
// }

func MapTo[T any, K comparable](records []T, keys []K, keyFn KeyFn[T, K]) []T {
	var m = make(map[K]T)
	for _, record := range records {
		m[keyFn(record)] = record
	}
	var result []T
	for _, key := range keys {
		if val, ok := m[key]; ok {
			result = append(result, val)
		}
	}

	return result
}

// export function mapToMany<R, K>(
//
//	records: ReadonlyArray<R>,
//	keys: ReadonlyArray<K>,
//	keyFn: (record: R) => K,
//
//	): Array<R[]> {
//	  const group = new Map<K, R[]>(keys.map((key) => [key, []]));
//	  records.forEach((record) => (group.get(keyFn(record)) || []).push(record));
//	  return Array.from(group.values());
//	}
func MapToMany[T any, K comparable](records []T, keys []K, keyFn KeyFn[T, K]) [][]T {
	var m = make(map[K][]T)
	for _, record := range records {
		m[keyFn(record)] = append(m[keyFn(record)], record)
	}
	var result [][]T
	for _, key := range keys {
		result = append(result, m[key])
	}

	return result
}
func Map[T1, T2 any](s []T1, f func(T1) T2) []T2 {
	r := make([]T2, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}

func Filter[T any](s []T, f func(T) bool) []T {
	var r []T
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func Reduce[T1, T2 any](s []T1, accumulator T2, f func(T2, T1) T2) T2 {
	r := accumulator
	for _, v := range s {
		r = f(r, v)
	}
	return r
}
