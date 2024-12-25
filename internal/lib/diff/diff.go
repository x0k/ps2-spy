package diff

import (
	"iter"
	"maps"
	"slices"
)

type Diff[T any] struct {
	ToAdd []T
	ToDel []T
}

func (d Diff[T]) IsEmpty() bool {
	return len(d.ToAdd) == 0 && len(d.ToDel) == 0
}

func mapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func diff[T comparable, V any](
	m map[T]V,
	new iter.Seq[T],
	newCount int,
) Diff[T] {
	toAdd := make([]T, 0, newCount)
	for k := range new {
		if _, ok := m[k]; ok {
			delete(m, k)
		} else {
			toAdd = append(toAdd, k)
		}
	}
	if len(m) == 0 {
		return Diff[T]{
			ToAdd: toAdd,
		}
	}
	return Diff[T]{
		ToAdd: toAdd,
		ToDel: mapKeys(m),
	}
}

func SlicesDiff[T comparable](old []T, new []T) Diff[T] {
	if len(old) == 0 {
		return Diff[T]{
			ToAdd: new,
		}
	}
	if len(new) == 0 {
		return Diff[T]{
			ToDel: old,
		}
	}
	m := make(map[T]struct{}, len(old))
	for _, v := range old {
		m[v] = struct{}{}
	}
	return diff(m, slices.Values(new), len(new))
}

func MapKeysDiff[K comparable, V any](old map[K]V, new map[K]V) Diff[K] {
	if len(old) == 0 {
		return Diff[K]{
			ToAdd: mapKeys(new),
		}
	}
	if len(new) == 0 {
		return Diff[K]{
			ToDel: mapKeys(old),
		}
	}
	return diff(maps.Clone(old), maps.Keys(new), len(new))
}
