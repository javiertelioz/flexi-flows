package workflow

import (
	"context"
	"fmt"
	"sync"
)

type BranchNode struct {
	Node[interface{}]
	Branches []NodeInterface
}

func (b *BranchNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(b.ID, b.Type, "context cancelled before branch execution", ctx.Err())
	default:
	}

	fmt.Printf("Executing BranchNode: %s\n", b.ID)

	if len(b.Branches) == 0 {
		return map[string]interface{}{
			"message": "no branches defined",
			"data":    data,
		}, nil
	}

	// Ejecutar todas las ramas en paralelo
	var wg sync.WaitGroup
	results := make([]interface{}, len(b.Branches))
	errors := make([]error, len(b.Branches))

	for i, branch := range b.Branches {
		wg.Add(1)
		go func(index int, branchNode NodeInterface) {
			defer wg.Done()

			fmt.Printf("Executing branch with ID: %s\n", branchNode.GetID())

			// Crear contexto derivado para cada rama
			branchCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			result, err := wm.ExecuteNodeWithContext(branchCtx, branchNode, data)
			results[index] = result
			errors[index] = err
		}(i, branch)
	}

	wg.Wait()

	// Verificar si hubo errores
	var firstError error
	successCount := 0
	for _, err := range errors {
		if err != nil {
			if firstError == nil {
				firstError = err
			}
		} else {
			successCount++
		}
	}

	// Si hay errores, devolver el primero con contexto adicional
	if firstError != nil {
		return nil, NewWorkflowError(b.ID, b.Type,
			fmt.Sprintf("%d of %d branches failed", len(b.Branches)-successCount, len(b.Branches)),
			firstError).
			WithContext("success_count", successCount).
			WithContext("total_count", len(b.Branches))
	}

	finalResult := map[string]interface{}{
		"results":       results,
		"success_count": successCount,
		"total_count":   len(b.Branches),
		"data":          data,
	}

	if len(b.Next) > 0 {
		return wm.ExecuteNodeWithContext(ctx, b.Next[0], finalResult)
	}

	return finalResult, nil
}
