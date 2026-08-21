package util

// FilterInPlace compacts items in place, reusing the underlying array for
// zero-allocation filtering. The returned slice shares memory with the input.
func FilterInPlace[T any](items []T, keep func(T) bool) []T {
	out := items[:0]
	for _, it := range items {
		if keep(it) {
			out = append(out, it)
		}
	}
	return out
}

// Take returns the first n items by reslicing the input, sharing its backing array.
func Take[T any](items []T, n int) []T {
	if n < 0 {
		n = 0
	}
	if len(items) > n {
		return items[:n]
	}
	return items
}
