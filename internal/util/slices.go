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

// Take returns a copy of the first n items. The result does not share memory
// with the input, so mutating it cannot corrupt the caller's slice.
func Take[T any](items []T, n int) []T {
	if n < 0 {
		n = 0
	}
	if n > len(items) {
		n = len(items)
	}
	out := make([]T, n)
	copy(out, items[:n])
	return out
}
