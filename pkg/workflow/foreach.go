package workflow

import (
	"context"
	"fmt"
	"reflect"
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

	// Convertir data a slice
	items, err := n.convertToSlice(data)
	if err != nil {
		return nil, NewWorkflowError(n.ID, n.Type,
			"failed to convert data to iterable collection", err)
	}

	if len(items) == 0 {
		return nil, NewWorkflowError(n.ID, n.Type,
			"no items to process in foreach", fmt.Errorf("empty collection"))
	}

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

// convertToSlice convierte los datos a un slice para poder iterar
func (n *ForeachNode) convertToSlice(data interface{}) ([]interface{}, error) {
	if slice, ok := data.([]interface{}); ok {
		return slice, nil
	}

	// Usar reflection para manejar diferentes tipos de slices
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Slice {
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
		}
		return result, nil
	}

	// Si es un solo elemento, crear slice con ese elemento
	return []interface{}{data}, nil
}
