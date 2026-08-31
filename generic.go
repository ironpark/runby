package runby

// The five detection axes are declared the same way — a table of products and
// a slice of drivers built from it — so the plumbing around them is written
// once here and instantiated per axis rather than copied.

// cloneSlice returns a copy of s, so that handing a built-in table to a
// caller — or to Detect, which sorts the agent axis — cannot reorder or
// truncate the table every other caller shares.
func cloneSlice[T any](s []T) []T {
	out := make([]T, len(s))
	copy(out, s)
	return out
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

// slug renders a product identity as its stable string, mapping the zero value
// to the axis's unknown constant so the output is never empty.
func slug[T ~string](value, unknown T) string {
	if value == "" {
		return string(unknown)
	}
	return string(value)
}
