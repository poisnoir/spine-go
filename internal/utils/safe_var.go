package utils

import "sync"

type SafeVar[T any] struct {
	data T
	mu   sync.RWMutex
}

func NewSafeVar[T any](data T) *SafeVar[T] {
	return &SafeVar[T]{data: data}
}

func (sv *SafeVar[T]) Get() T {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.data
}

func (sv *SafeVar[T]) Set(data T) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	sv.data = data
}
