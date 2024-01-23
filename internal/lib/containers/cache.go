package containers

type Cache[T any] interface {
	Get() (T, bool)
	Set(value T)
}

type KeyedCache[K comparable, T any] interface {
	Get(key K) (T, bool)
	// Returns bool to compatibility with expirable.LRU
	// eviction indicator
	Add(key K, value T) bool
}
