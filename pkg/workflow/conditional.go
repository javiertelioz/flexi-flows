package workflow

import (
	"log"
)

type ConditionalNode[T any] struct {
	Node[T]
	ConditionFunc func(T) (T, error)
	TrueNext      NodeInterface[T]
	FalseNext     NodeInterface[T]
}

func (n *ConditionalNode[T]) Execute(wm *WorkflowManager[T], data T) (T, error) {
	result, err := n.ConditionFunc(data)
	if err != nil {
		log.Printf("Error in condition function: %v", err)
		return data, err
	}

	// Attempt to convert result to bool
	conditionResult, ok := any(result).(bool)
	if !ok {
		log.Printf("Condition function did not return a bool")
		return data, nil
	}

	if conditionResult {
		if n.TrueNext != nil {
			return wm.ExecuteNode(n.TrueNext, data)
		}
	} else {
		if n.FalseNext != nil {
			return wm.ExecuteNode(n.FalseNext, data)
		}
	}
	return data, nil
}
