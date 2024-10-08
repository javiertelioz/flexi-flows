package workflow

import (
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

func (n *Node[T]) Execute(wm *WorkflowManager, data interface{}) (interface{}, error) {
	typedData, ok := data.(T)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected %T, got %T", typedData, data)
	}

	var err error

	if n.BeforeExecute != nil {
		typedData, err = n.BeforeExecute(typedData)
		if err != nil {
			return nil, err
		}
	}

	var result T
	if n.TaskFunc != nil {
		result, err = n.TaskFunc(typedData)
		if err != nil {
			return nil, err
		}
	}

	if n.AfterExecute != nil {
		result, err = n.AfterExecute(result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
