package common

import (
	"sync"
)

type ConcurrentMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		m: make(map[K]V),
	}
}

func (cm *ConcurrentMap[K, V]) Load(key K) (V, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	val, ok := cm.m[key]
	return val, ok
}

func (cm *ConcurrentMap[K, V]) Store(key K, value V) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.m[key] = value
}
