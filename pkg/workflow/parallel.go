package workflow

import (
	"sync"
)

type ParallelNode struct {
	Node
	ParallelTasks []NodeInterface
}

func (p *ParallelNode) Execute(wm *WorkflowManager, data interface{}) (interface{}, error) {
	var wg sync.WaitGroup
	results := make([]interface{}, len(p.ParallelTasks))
	errors := make([]error, len(p.ParallelTasks))

	for i, task := range p.ParallelTasks {
		wg.Add(1)
		go func(i int, task NodeInterface) {
			defer wg.Done()
			result, err := wm.executeNode(task, data)
			results[i] = result
			errors[i] = err
		}(i, task)
	}

	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}
