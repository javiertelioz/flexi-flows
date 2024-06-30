package workflow

type ExecutableNode interface {
	Execute(wm *WorkflowManager) error
}
