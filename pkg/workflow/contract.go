package workflow

type NodeInterface[T any] interface {
	GetID() string
	GetType() NodeType
	Execute(wm *WorkflowManager[T], data T) (T, error)
}
