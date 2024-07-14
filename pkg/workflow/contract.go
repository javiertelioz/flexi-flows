package workflow

type NodeInterface interface {
	GetID() string
	GetType() NodeType
	Execute(wm *WorkflowManager, data interface{}) (interface{}, error)
}
