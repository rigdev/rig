package utils

func InsertInList[T any](existing []T, obj T, equal func(T, T) bool) []T {
	for idx, o := range existing {
		if equal(o, obj) {
			existing[idx] = obj
			return existing
		}
	}
	return append(existing, obj)
}

func RemoveFromList[T any, K any](existing []T, key K, equal func(T, K) bool) []T {
	for idx, o := range existing {
		if equal(o, key) {
			return append(existing[:idx], existing[idx+1:]...)
		}
	}
	return existing
}
