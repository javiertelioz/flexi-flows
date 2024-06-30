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
	Execute(wm *WorkflowManager, data interface{}) (interface{}, error)
}

type Node struct {
	ID            string
	Type          NodeType
	TaskFunc      func(interface{}) (interface{}, error)
	SubDag        *Graph
	Next          []NodeInterface
	BeforeExecute func()
	AfterExecute  func()
}

func (n *Node) GetID() string {
	return n.ID
}

func (n *Node) GetType() NodeType {
	return n.Type
}

func (n *Node) Execute(wm *WorkflowManager, data interface{}) (interface{}, error) {
	if n.BeforeExecute != nil {
		n.BeforeExecute()
	}

	var result interface{}
	var err error
	if n.TaskFunc != nil {
		result, err = n.TaskFunc(data)
		if err != nil {
			return nil, err
		}
	}

	if n.AfterExecute != nil {
		n.AfterExecute()
	}

	return result, nil
}
