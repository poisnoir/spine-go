package utils

import "sync"

type SafeCounter struct {
	count int
	mu    sync.RWMutex
}

func NewSafeCounter(start int) *SafeCounter {
	return &SafeCounter{count: start}
}

func (sc *SafeCounter) Get() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.count
}

func (sc *SafeCounter) Add() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.count = sc.count + 1
}

func (sc *SafeCounter) Sub() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.count = sc.count - 1
}
