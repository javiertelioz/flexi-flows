package workflow

type ForeachNode struct {
	ID          string
	Type        NodeType
	Collection  []interface{}
	IterateFunc func(interface{}) error
	Next        []NodeInterface
}

func (n *ForeachNode) Execute(wm *WorkflowManager) error {
	for _, item := range n.Collection {
		if err := n.IterateFunc(item); err != nil {
			return err
		}
	}
	if len(n.Next) > 0 {
		return wm.executeNode(n.Next[0])
	}
	return nil
}

func (n *ForeachNode) GetID() string {
	return n.ID
}

func (n *ForeachNode) GetType() NodeType {
	return n.Type
}
