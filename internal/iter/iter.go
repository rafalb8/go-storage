package iter

type MapFunc[V, O any] func(x V) O

// MapSlice creates a new slice from calling a function for every slice element
func MapSlice[V any, O any](val []V, fn MapFunc[V, O]) []O {
	newValues := make([]O, len(val))
	for i, v := range val {
		newValues[i] = fn(v)
	}
	return newValues
}

// Map creates a new map from calling a function for every map element
func MapMap[K comparable, V, O any](val map[K]V, fn MapFunc[V, O]) map[K]O {
	newValues := make(map[K]O, len(val))
	for k, v := range val {
		newValues[k] = fn(v)
	}
	return newValues
}