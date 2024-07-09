package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type JSONStateStore struct {
	filePath string
	mu       sync.Mutex
}

func NewJSONStateStore(filePath string) *JSONStateStore {
	return &JSONStateStore{
		filePath: filePath,
	}
}

func (s *JSONStateStore) SaveState(nodeID string, data interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	state := make(map[string]interface{})
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&state); err != nil && !errors.Is(err, os.ErrNotExist) && err.Error() != "EOF" {
		return err
	}

	state[nodeID] = data

	file.Seek(0, 0)
	file.Truncate(0)
	encoder := json.NewEncoder(file)
	return encoder.Encode(state)
}

func (s *JSONStateStore) LoadState(nodeID string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	state := make(map[string]interface{})
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&state); err != nil && err.Error() != "EOF" {
		return nil, err
	}

	data, exists := state[nodeID]
	if !exists {
		return nil, nil
	}
	return data, nil
}
