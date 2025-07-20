package workflow

import (
	"fmt"
	"runtime"
	"time"
)

// NodeType define los tipos de nodos disponibles en el workflow
// Mantiene compatibilidad con el sistema anterior usando int
type NodeType int

const (
	Task NodeType = iota
	SubDag
	Conditional
	Foreach
	Branch
	// Nuevos tipos de nodos
	Parallel
	HTTP
	Delay
	Transform
	Validation
	Merge
	Split
	Filter
)

// Alias para mantener compatibilidad y semántica mejorada
const (
	Decision = Branch  // Alias para Branch
	Loop     = Foreach // Alias para Foreach
	Subflow  = SubDag  // Alias para SubDag
)

// String implementa el interfaz Stringer para NodeType
func (nt NodeType) String() string {
	switch nt {
	case Task:
		return "task"
	case SubDag:
		return "subdag"
	case Conditional:
		return "conditional"
	case Foreach:
		return "foreach"
	case Branch:
		return "branch"
	case Parallel:
		return "parallel"
	case HTTP:
		return "http"
	case Delay:
		return "delay"
	case Transform:
		return "transform"
	case Validation:
		return "validation"
	case Merge:
		return "merge"
	case Split:
		return "split"
	case Filter:
		return "filter"
	default:
		return "unknown"
	}
}

// IsValid verifica si el tipo de nodo es válido
func (nt NodeType) IsValid() bool {
	return nt >= Task && nt <= Filter
}

// HookType define los tipos de hooks disponibles
type HookType string

const (
	BeforeExecution HookType = "before"
	AfterExecution  HookType = "after"
	OnSuccess       HookType = "success"
	OnError         HookType = "error"
	OnComplete      HookType = "complete"
)

// WorkflowError representa un error estructurado del workflow
type WorkflowError struct {
	NodeID    string                 `json:"node_id"`
	NodeType  NodeType               `json:"node_type"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Stack     []string               `json:"stack,omitempty"`
	Cause     error                  `json:"-"`
}

// Error implementa el interfaz error
func (we *WorkflowError) Error() string {
	return fmt.Sprintf("workflow error in node %s (%s): %s", we.NodeID, we.NodeType, we.Message)
}

// Unwrap permite usar errors.Is y errors.As
func (we *WorkflowError) Unwrap() error {
	return we.Cause
}

// NewWorkflowError crea un nuevo error de workflow con stack trace
func NewWorkflowError(nodeID string, nodeType NodeType, message string, cause error) *WorkflowError {
	stack := make([]string, 0, 10)
	for i := 1; i < 10; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stack = append(stack, fmt.Sprintf("%s:%d", file, line))
	}

	return &WorkflowError{
		NodeID:    nodeID,
		NodeType:  nodeType,
		Message:   message,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
		Stack:     stack,
		Cause:     cause,
	}
}

// WithContext añade contexto adicional al error
func (we *WorkflowError) WithContext(key string, value interface{}) *WorkflowError {
	if we.Context == nil {
		we.Context = make(map[string]interface{})
	}
	we.Context[key] = value
	return we
}

// ExecutionContext representa el contexto de ejecución de un nodo
type ExecutionContext struct {
	WorkflowID  string                 `json:"workflow_id"`
	NodeID      string                 `json:"node_id"`
	ExecutionID string                 `json:"execution_id"`
	Data        interface{}            `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
}

// HookContext representa el contexto disponible para los hooks
type HookContext struct {
	NodeID      string                 `json:"node_id"`
	NodeType    NodeType               `json:"node_type"`
	HookType    HookType               `json:"hook_type"`
	Data        interface{}            `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	ExecutionID string                 `json:"execution_id"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NodeState representa el estado de ejecución de un nodo
type NodeState string

const (
	StateReady     NodeState = "ready"
	StateRunning   NodeState = "running"
	StateCompleted NodeState = "completed"
	StateFailed    NodeState = "failed"
	StateSkipped   NodeState = "skipped"
)

// WorkflowState representa el estado general del workflow
type WorkflowState string

const (
	WorkflowStateReady     WorkflowState = "ready"
	WorkflowStateRunning   WorkflowState = "running"
	WorkflowStateCompleted WorkflowState = "completed"
	WorkflowStateFailed    WorkflowState = "failed"
	WorkflowStatePaused    WorkflowState = "paused"
)

// ValidationRule representa una regla de validación
type ValidationRule struct {
	Field    string      `json:"field" yaml:"field"`
	Type     string      `json:"type" yaml:"type"`
	Required bool        `json:"required,omitempty" yaml:"required,omitempty"`
	Min      interface{} `json:"min,omitempty" yaml:"min,omitempty"`
	Max      interface{} `json:"max,omitempty" yaml:"max,omitempty"`
	Pattern  string      `json:"pattern,omitempty" yaml:"pattern,omitempty"`
}

// TaskFunc define la firma de una función de tarea
type TaskFunc func(data map[string]interface{}) (map[string]interface{}, error)

// HookFunc define la firma de una función de hook simple
type HookFunc func() error

// HookFuncWithData define la firma de una función de hook con datos
type HookFuncWithData func(data interface{}) error

// HookFuncWithContext define la firma de una función de hook con contexto completo
type HookFuncWithContext func(ctx *HookContext) error
