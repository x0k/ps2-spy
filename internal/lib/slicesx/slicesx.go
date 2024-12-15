package slicesx

import "slices"

func GroupBy[A any, K comparable](input []A, f func(A) K) map[K][]A {
	result := make(map[K][]A)
	for _, v := range input {
		key := f(v)
		result[key] = append(result[key], v)
	}
	return result
}

func Filter[A any](arr []A, filter func(index int) bool) []A {
	shift := 0
	clone := slices.Clone(arr)
	for i := 0; i < len(clone); i++ {
		if filter(i) {
			clone[i-shift] = clone[i]
		} else {
			shift++
		}
	}
	return clone[:len(clone)-shift]
}
