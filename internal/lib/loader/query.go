package loader

type Query[K comparable] struct {
	Provider string
	Key      K
}

func NewQuery[K comparable](provider string, key K) Query[K] {
	return Query[K]{
		Provider: provider,
		Key:      key,
	}
}
