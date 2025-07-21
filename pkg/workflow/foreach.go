package workflow

import (
	"context"
	"fmt"
)

type ForeachNode struct {
	Node[interface{}]
	Collection  []interface{}
	IterateFunc func(interface{}) (interface{}, error)
}

func (n *ForeachNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado antes de comenzar
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(n.ID, n.Type, "context cancelled before foreach execution", ctx.Err())
	default:
	}

	var results []interface{}

	for i, item := range n.Collection {
		// Verificar cancelación en cada iteración
		select {
		case <-ctx.Done():
			return nil, NewWorkflowError(n.ID, n.Type,
				fmt.Sprintf("context cancelled at iteration %d", i), ctx.Err())
		default:
		}

		result, err := n.IterateFunc(item)
		if err != nil {
			return nil, NewWorkflowError(n.ID, n.Type,
				fmt.Sprintf("iteration %d failed", i), err).
				WithContext("iteration_index", i).
				WithContext("item", item)
		}
		results = append(results, result)
	}

	// Resultado combinado con todos los resultados de las iteraciones
	finalResult := map[string]interface{}{
		"results":       results,
		"count":         len(results),
		"original_data": data,
	}

	if len(n.Next) > 0 {
		return wm.ExecuteNodeWithContext(ctx, n.Next[0], finalResult)
	}

	return finalResult, nil
}
