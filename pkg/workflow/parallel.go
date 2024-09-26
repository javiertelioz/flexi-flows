package workflow

import (
	"fmt"
	"sync"
)

type ParallelNode[T any] struct {
	Node[T]
	ParallelTasks []NodeInterface[T]
}

func (p *ParallelNode[T]) Execute(wm *WorkflowManager[T], data T) (T, error) {
	var wg sync.WaitGroup
	results := make(map[string]T)
	errorsMap := make(map[string]error)
	mu := sync.Mutex{}

	for _, task := range p.ParallelTasks {
		wg.Add(1)
		go func(task NodeInterface[T]) {
			defer wg.Done()
			result, err := wm.ExecuteNode(task, data)
			mu.Lock()
			defer mu.Unlock()
			results[task.GetID()] = result
			if err != nil {
				errorsMap[task.GetID()] = err
			}
		}(task)
	}

	wg.Wait()

	if len(errorsMap) > 0 {
		return data, fmt.Errorf("errors in parallel tasks: %v", errorsMap)
	}

	// The result is a map of results
	var combinedResult T
	if res, ok := any(results).(T); ok {
		combinedResult = res
	} else {
		combinedResult = any(results).(T)
	}

	// If there is a next node, execute it with the combined result
	if len(p.Next) > 0 {
		return wm.ExecuteNode(p.Next[0], combinedResult)
	}

	return combinedResult, nil
}
