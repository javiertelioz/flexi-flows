package workflow

import (
	"sync"
)

type MemoryStateStore struct {
	state map[string]interface{}
	mu    sync.Mutex
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		state: make(map[string]interface{}),
	}
}

func (s *MemoryStateStore) SaveState(nodeID string, data interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[nodeID] = data
	return nil
}

func (s *MemoryStateStore) LoadState(nodeID string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, exists := s.state[nodeID]
	if !exists {
		return nil, nil
	}
	return data, nil
}
