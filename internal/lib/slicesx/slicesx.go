package slicesx

import (
	"slices"
)

func Map[T any, R any](slice []T, f func(T) R) []R {
	result := make([]R, 0, len(slice))
	for _, v := range slice {
		result = append(result, f(v))
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
