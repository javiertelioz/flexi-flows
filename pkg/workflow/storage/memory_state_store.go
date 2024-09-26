package storage

import (
	"sync"
)

type MemoryStateStore[T any] struct {
	state map[string]T
	mu    sync.Mutex
}

func NewMemoryStateStore[T any]() *MemoryStateStore[T] {
	return &MemoryStateStore[T]{
		state: make(map[string]T),
	}
}

func (s *MemoryStateStore[T]) SaveState(nodeID string, data T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[nodeID] = data
	return nil
}

func (s *MemoryStateStore[T]) LoadState(nodeID string) (T, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, exists := s.state[nodeID]
	return data, exists, nil
}
