package utils

import "sync"

type SafeArr[T any] struct {
	data []T
	mu   sync.RWMutex
}

func NewSafeArr[T any]() *SafeArr[T] {
	return &SafeArr[T]{
		data: []T{},
	}
}

func (sv *SafeArr[T]) Get() []T {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	cp := make([]T, len(sv.data))
	copy(cp, sv.data)
	return cp
}

func (sv *SafeArr[T]) Add(data T) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	sv.data = append(sv.data, data)
}

func (sv *SafeArr[T]) Remove(index int) {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	if index < 0 || index >= len(sv.data) {
		return
	}
	sv.data = append(sv.data[:index], sv.data[index+1:]...)
}
