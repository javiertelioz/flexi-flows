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
	BeforeExecute func(interface{}) (interface{}, error)
	AfterExecute  func(interface{}) (interface{}, error)
}

func (n *Node) GetID() string {
	return n.ID
}

func (n *Node) GetType() NodeType {
	return n.Type
}

func (n *Node) Execute(wm *WorkflowManager, data interface{}) (interface{}, error) {
	var err error

	if n.BeforeExecute != nil {
		data, err = n.BeforeExecute(data)
		if err != nil {
			return nil, err
		}
	}

	var result interface{}
	if n.TaskFunc != nil {
		result, err = n.TaskFunc(data)
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
