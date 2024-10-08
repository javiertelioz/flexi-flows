package workflow

type ConditionalNode struct {
	Node[interface{}]
	Condition func(data interface{}) bool
	TrueNext  NodeInterface
	FalseNext NodeInterface
}

func (n *ConditionalNode) Execute(wm *WorkflowManager, data interface{}) (interface{}, error) {
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
