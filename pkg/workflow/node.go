package workflow

import (
	"fmt"
)

type Node[T any] struct {
	ID            string
	Type          NodeType
	TaskFunc      func(T) (T, error)
	SubDag        *Graph[T]
	Next          []NodeInterface[T]
	BeforeExecute func(T) (T, error)
	AfterExecute  func(T) (T, error)
}

func (n *Node[T]) GetID() string {
	return n.ID
}

func (n *Node[T]) GetType() NodeType {
	return n.Type
}

func (n *Node[T]) Execute(wm *WorkflowManager[T], data T) (T, error) {
	var err error

	// Input data validation
	if any(data) == nil {
		return data, fmt.Errorf("input data is nil for node %s", n.ID)
	}

	// Execute before hook if available
	if n.BeforeExecute != nil {
		data, err = n.BeforeExecute(data)
		if err != nil {
			return data, err
		}
	}

	var result T

	// Execute main task function if available
	if n.TaskFunc != nil {
		result, err = n.TaskFunc(data)
		if err != nil {
			return result, err
		}
	} else {
		result = data
	}

	// Execute after hook if available
	if n.AfterExecute != nil {
		result, err = n.AfterExecute(result)
		if err != nil {
			return result, err
		}
	}

	// If there is a next node, execute it with the current result
	if len(n.Next) > 0 {
		return wm.ExecuteNode(n.Next[0], result)
	}

	return result, nil
}
