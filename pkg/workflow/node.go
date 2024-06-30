package workflow

type NodeType int

const (
	Task NodeType = iota
	SubDag
	Conditional
	Foreach
	Branch
)

type NodeInterface interface {
	GetID() string
	GetType() NodeType
	Execute(wm *WorkflowManager) error
}

type Node struct {
	ID       string
	Type     NodeType
	TaskFunc func() error
	SubDag   *Graph
	Next     []NodeInterface
}

func (n *Node) GetID() string {
	return n.ID
}

func (n *Node) GetType() NodeType {
	return n.Type
}

func (n *Node) Execute(wm *WorkflowManager) error {
	if n.TaskFunc != nil {
		return n.TaskFunc()
	}
	return nil
}
