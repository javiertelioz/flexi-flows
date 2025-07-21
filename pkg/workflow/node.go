package workflow

import (
	"context"
	"fmt"
)

type Node[T any] struct {
	ID            string
	Type          NodeType
	TaskFunc      func(T) (T, error)
	SubDag        *Graph
	Next          []NodeInterface
	BeforeExecute func(T) (T, error)
	AfterExecute  func(T) (T, error)
}

func (n *Node[T]) GetID() string {
	return n.ID
}

func (n *Node[T]) GetType() NodeType {
	return n.Type
}

func (n *Node[T]) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(n.ID, n.Type, "context cancelled before node execution", ctx.Err())
	default:
	}

	typedData, ok := data.(T)
	if !ok {
		return nil, NewWorkflowError(n.ID, n.Type, "invalid data type",
			fmt.Errorf("expected %T, got %T", typedData, data))
	}

	var err error

	if n.BeforeExecute != nil {
		typedData, err = n.BeforeExecute(typedData)
		if err != nil {
			return nil, NewWorkflowError(n.ID, n.Type, "before execute failed", err)
		}
	}

	var result T
	if n.TaskFunc != nil {
		result, err = n.TaskFunc(typedData)
		if err != nil {
			return nil, NewWorkflowError(n.ID, n.Type, "task execution failed", err)
		}
	} else {
		result = typedData
	}

	if n.AfterExecute != nil {
		result, err = n.AfterExecute(result)
		if err != nil {
			return nil, NewWorkflowError(n.ID, n.Type, "after execute failed", err)
		}
	}

	if len(n.Next) > 0 {
		return wm.ExecuteNodeWithContext(ctx, n.Next[0], result)
	}

	return result, nil
}
