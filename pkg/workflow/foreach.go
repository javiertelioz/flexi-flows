package workflow

import (
	"fmt"
	"sync"
)

type ExecutionMode string

const (
	Synchronous  ExecutionMode = "synchronous"
	Asynchronous ExecutionMode = "asynchronous"
)

type ForeachNode[T any] struct {
	Node[T]
	IterateFunc   func(T) (T, error)
	ExecutionMode ExecutionMode
}

func (n *ForeachNode[T]) Execute(wm *WorkflowManager[T], data T) (T, error) {
	// Se espera que 'data' sea una colección []T
	collection, ok := any(data).([]T)
	if !ok {
		return data, fmt.Errorf("ForeachNode expects data of type []T, got %T", data)
	}

	var results []T
	var errorsSlice []error

	if n.ExecutionMode == Asynchronous {
		var wg sync.WaitGroup
		results = make([]T, 0, len(collection))
		errorsSlice = make([]error, 0)
		mu := sync.Mutex{}

		for _, item := range collection {
			wg.Add(1)
			go func(item T) {
				defer wg.Done()
				result, err := n.IterateFunc(item)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					errorsSlice = append(errorsSlice, err)
				} else {
					results = append(results, result)
				}
			}(item)
		}

		wg.Wait()
	} else {
		// Ejecución síncrona
		results = make([]T, 0, len(collection))
		errorsSlice = make([]error, 0)

		for _, item := range collection {
			result, err := n.IterateFunc(item)
			if err != nil {
				errorsSlice = append(errorsSlice, err)
			} else {
				results = append(results, result)
			}
		}
	}

	if len(errorsSlice) > 0 {
		return data, fmt.Errorf("errors in foreach tasks: %v", errorsSlice)
	}

	if len(n.Next) > 0 {
		if res, ok := any(results).(T); ok {
			return wm.ExecuteNode(n.Next[0], res)
		} else {
			return wm.ExecuteNode(n.Next[0], any(results).(T))
		}
	}

	return data, nil
}
