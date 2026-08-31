package runby

// The four detection axes are declared the same way — a table of products, a
// slice of drivers built from it, and options that add to or replace those
// drivers — so the plumbing around them is written once here and instantiated
// per axis rather than copied.

// cloneSlice returns a copy of s. The exported XDrivers accessors use it so a
// caller can reorder or filter the built-in tables without affecting them.
func cloneSlice[T any](s []T) []T {
	out := make([]T, len(s))
	copy(out, s)
	return out
}

// replaceDrivers discards the drivers already configured. Passing none
// disables the axis.
func replaceDrivers[T any](dst *[]T, add []T) {
	*dst = cloneSlice(add)
}

// mapSlice applies fn to every element. It builds a driver table from a spec
// table, and an identity list from a driver table.
func mapSlice[T, R any](s []T, fn func(T) R) []R {
	out := make([]R, 0, len(s))
	for _, item := range s {
		out = append(out, fn(item))
	}
	return out
}

// indexBy builds a lookup from a driver table, so that a product is registered
// in exactly one place and its facts are derived rather than restated.
func indexBy[T any, K comparable, V any](s []T, fn func(T) (K, V)) map[K]V {
	out := make(map[K]V, len(s))
	for _, item := range s {
		key, value := fn(item)
		out[key] = value
	}
	return out
}

// lookupOr returns the value stored under key, or fallback when it is absent.
// Every axis uses it to answer a question about a product it may not know.
func lookupOr[K comparable, V any](m map[K]V, key K, fallback V) V {
	if value, ok := m[key]; ok {
		return value
	}
	return fallback
}

// slug renders a product identity as its stable string, mapping the zero value
// to the axis's unknown constant so the output is never empty.
func slug[T ~string](value, unknown T) string {
	if value == "" {
		return string(unknown)
	}
	return string(value)
}
