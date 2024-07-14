package nodes

import "github.com/javiertelioz/go-flows/pkg/workflow"

type ConditionalNode struct {
	workflow.Node[interface{}]
	Condition func(data interface{}) bool
	TrueNext  workflow.NodeInterface
	FalseNext workflow.NodeInterface
}

func (n *ConditionalNode) Execute(wm *workflow.WorkflowManager, data interface{}) (interface{}, error) {
	if n.Condition(data) {
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
