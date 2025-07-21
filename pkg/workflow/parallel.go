package workflow

import (
	"context"
	"fmt"
	"sync"
)

type ParallelNode struct {
	Node[interface{}]
	ParallelTasks []NodeInterface
}

func (n *ParallelNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(n.ID, n.Type, "context cancelled before parallel execution", ctx.Err())
	default:
	}

	if len(n.ParallelTasks) == 0 {
		return map[string]interface{}{
			"results": []interface{}{},
			"count":   0,
			"message": "no parallel tasks defined",
			"data":    data,
		}, nil
	}

	// Canal para recopilar resultados
	resultChan := make(chan struct {
		index  int
		result interface{}
		err    error
	}, len(n.ParallelTasks))

	// Usar WaitGroup para sincronización
	var wg sync.WaitGroup

	// Ejecutar cada tarea en paralelo
	for i, task := range n.ParallelTasks {
		wg.Add(1)
		go func(index int, node NodeInterface) {
			defer wg.Done()

			// Crear un contexto derivado para esta tarea
			taskCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			result, err := wm.ExecuteNodeWithContext(taskCtx, node, data)

			resultChan <- struct {
				index  int
				result interface{}
				err    error
			}{index: index, result: result, err: err}
		}(i, task)
	}

	// Goroutine para cerrar el canal cuando todas las tareas terminen
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Recopilar resultados
	results := make([]interface{}, len(n.ParallelTasks))
	var errors []error
	completed := 0

	for taskResult := range resultChan {
		select {
		case <-ctx.Done():
			return nil, NewWorkflowError(n.ID, n.Type, "context cancelled during parallel execution", ctx.Err())
		default:
		}

		completed++
		if taskResult.err != nil {
			errors = append(errors, fmt.Errorf("task %d failed: %w", taskResult.index, taskResult.err))
		} else {
			results[taskResult.index] = taskResult.result
		}
	}

	// Si hay errores, devolver el primero pero incluir información de todos
	if len(errors) > 0 {
		return nil, NewWorkflowError(n.ID, n.Type,
			fmt.Sprintf("%d of %d parallel tasks failed", len(errors), len(n.ParallelTasks)),
			errors[0]).
			WithContext("error_count", len(errors)).
			WithContext("completed_count", completed-len(errors)).
			WithContext("total_count", len(n.ParallelTasks))
	}

	finalResult := map[string]interface{}{
		"results":    results,
		"count":      len(results),
		"completed":  completed,
		"data":       data,
	}

	if len(n.Next) > 0 {
		return wm.ExecuteNodeWithContext(ctx, n.Next[0], finalResult)
	}

	return finalResult, nil
}
