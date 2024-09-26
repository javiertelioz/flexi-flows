package workflow

import (
	"fmt"
	"sync"
)

type BranchNode[T any] struct {
	Node[T]
	Branches []NodeInterface[T]
}

func (b *BranchNode[T]) Execute(wm *WorkflowManager[T], data T) (T, error) {
	var wg sync.WaitGroup
	results := make(map[string]T)
	errorsMap := make(map[string]error)
	mu := sync.Mutex{}

	for _, branch := range b.Branches {
		wg.Add(1)
		go func(branch NodeInterface[T]) {
			defer wg.Done()
			result, err := wm.ExecuteNode(branch, data)
			mu.Lock()
			defer mu.Unlock()
			results[branch.GetID()] = result
			if err != nil {
				errorsMap[branch.GetID()] = err
			}
		}(branch)
	}

	wg.Wait()

	if len(errorsMap) > 0 {
		return data, fmt.Errorf("errors in branch tasks: %v", errorsMap)
	}

	// You can decide how to combine branch results here

	// If there is a next node, execute it
	if len(b.Next) > 0 {
		// For simplicity, we pass the original data
		return wm.ExecuteNode(b.Next[0], data)
	}

	return data, nil
}
