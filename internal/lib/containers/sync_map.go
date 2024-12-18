package containers

import "sync"

type SyncMap[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]V
}

func NewSyncMap[K comparable, V any](m map[K]V) *SyncMap[K, V] {
	return &SyncMap[K, V]{m: m}
}

func (s *SyncMap[K, V]) Set(key K, value V) {
	s.mu.Lock()
	s.m[key] = value
	s.mu.Unlock()
}

func (s *SyncMap[K, V]) Get(key K) V {
	s.mu.Lock()
	v := s.m[key]
	s.mu.Unlock()
	return v
}

func (s *SyncMap[K, V]) Delete(key K) {
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()
}

func (s *SyncMap[K, V]) Pop(key K) V {
	s.mu.Lock()
	v := s.m[key]
	delete(s.m, key)
	s.mu.Unlock()
	return v
}
