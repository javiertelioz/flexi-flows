package workflow

import "context"

type NodeInterface interface {
	GetID() string
	GetType() NodeType
	Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error)
}

// Hook representa una función de hook que puede ser ejecutada
type Hook interface {
	Execute(ctx context.Context, hookCtx *HookContext) error
}

// HookManager maneja la ejecución de hooks
type HookManager interface {
	RegisterHook(hookType HookType, hook Hook)
	ExecuteHooks(ctx context.Context, hookType HookType, hookCtx *HookContext) error
}

// ExecutionResult representa el resultado de una ejecución
type ExecutionResult struct {
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
	Duration  int64                  `json:"duration_ms"`
	Success   bool                   `json:"success"`
	Error     error                  `json:"-"`
	NodeID    string                 `json:"node_id"`
	Timestamp int64                  `json:"timestamp"`
}
