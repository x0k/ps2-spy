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

func mapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func newKeysMap[T comparable](keys []T) map[T]struct{} {
	m := make(map[T]struct{}, len(keys))
	for _, k := range keys {
		m[k] = struct{}{}
	}
	return m
}

func diff[T comparable, V any](
	old map[T]V,
	new iter.Seq[T],
	newCount int,
) Diff[T] {
	toAdd := make([]T, 0, newCount)
	for k := range new {
		if _, ok := old[k]; ok {
			delete(old, k)
		} else {
			toAdd = append(toAdd, k)
		}
	}
	if len(old) == 0 {
		return Diff[T]{
			ToAdd: toAdd,
		}
	}
	return Diff[T]{
		ToAdd: toAdd,
		ToDel: mapKeys(old),
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
	return diff(newKeysMap(old), slices.Values(new), len(new))
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

func SliceAndMapValuesDiff[T comparable, K comparable](
	old []T,
	new map[K]T,
) Diff[T] {
	if len(old) == 0 {
		return Diff[T]{
			ToAdd: mapValues(new),
		}
	}
	if len(new) == 0 {
		return Diff[T]{
			ToDel: old,
		}
	}
	return diff(newKeysMap(old), maps.Values(new), len(new))
}

func MissingKeys[K comparable, V any](m map[K]V, keys []K) []K {
	mLen := len(m)
	kLen := len(keys)
	if mLen == kLen {
		return nil
	}
	if mLen == 0 {
		return keys
	}
	result := make([]K, 0, kLen-mLen)
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			result = append(result, k)
		}
	}
	return result
}
