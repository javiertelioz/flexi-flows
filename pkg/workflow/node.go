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
	var err error
	typedData := data.(T)

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
