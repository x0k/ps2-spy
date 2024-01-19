package diff

type Diff[T any] struct {
	ToAdd []T
	ToDel []T
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
	toAdd := make([]T, 0, len(new))
	for _, v := range new {
		if _, ok := m[v]; ok {
			delete(m, v)
		} else {
			toAdd = append(toAdd, v)
		}
	}
	if len(m) == 0 {
		return Diff[T]{
			ToAdd: toAdd,
		}
	}
	toDel := make([]T, 0, len(old))
	for k := range m {
		toDel = append(toDel, k)
	}
	return Diff[T]{toAdd, toDel}
}
