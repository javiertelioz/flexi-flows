package workflow

type ForeachNode struct {
	Node[interface{}]
	Collection  []interface{}
	IterateFunc func(interface{}) (interface{}, error)
}

func (n *ForeachNode) Execute(wm *WorkflowManager, data interface{}) (interface{}, error) {
	for _, item := range n.Collection {
		_, err := n.IterateFunc(item)
		if err != nil {
			return nil, err
		}
	}
	if len(n.Next) > 0 {
		return wm.ExecuteNode(n.Next[0], data)
	}
	return nil, nil
}
