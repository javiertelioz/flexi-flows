package workflow

import (
	"context"
	"fmt"
	"runtime"
	"strings"
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

// HookType define los tipos de hooks disponibles
type HookType int

const (
	PreExecution HookType = iota
	PostExecution
	OnError
	OnSuccess
	OnCompletion
	PreNode
	PostNode
	BeforeExecution // Alias para PreExecution
	AfterExecution  // Alias para PostExecution
	OnComplete      // Alias para OnCompletion
)

// String implementa el interfaz Stringer para HookType
func (ht HookType) String() string {
	switch ht {
	case PreExecution:
		return "pre_execution"
	case PostExecution:
		return "post_execution"
	case OnError:
		return "on_error"
	case OnSuccess:
		return "on_success"
	case OnCompletion:
		return "on_completion"
	case PreNode:
		return "pre_node"
	case PostNode:
		return "post_node"
	default:
		return "unknown"
	}
}

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

// TaskFunc define la firma de función para tareas
type TaskFunc func(context.Context, interface{}) (interface{}, error)

// HookFunc define la firma de función para hooks
type HookFunc func(HookContext) error

// NodeConfig representa la configuración de un nodo
type NodeConfig struct {
	ID           string                 `json:"id" yaml:"id"`
	Name         string                 `json:"name" yaml:"name"`
	Type         NodeType               `json:"type" yaml:"type"`
	Function     string                 `json:"function,omitempty" yaml:"function,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Navigation   map[string]string      `json:"navigation,omitempty" yaml:"navigation,omitempty"`
	Settings     map[string]interface{} `json:"settings,omitempty" yaml:"settings,omitempty"`

	// Campos específicos para diferentes tipos de nodos
	URL        string                 `json:"url,omitempty" yaml:"url,omitempty"`
	Method     string                 `json:"method,omitempty" yaml:"method,omitempty"`
	Headers    map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body       interface{}            `json:"body,omitempty" yaml:"body,omitempty"`
	Duration   string                 `json:"duration,omitempty" yaml:"duration,omitempty"`
	Condition  string                 `json:"condition,omitempty" yaml:"condition,omitempty"`
	Collection string                 `json:"collection,omitempty" yaml:"collection,omitempty"`
	Rules      []ValidationRule       `json:"rules,omitempty" yaml:"rules,omitempty"`
	Tasks      []string               `json:"tasks,omitempty" yaml:"tasks,omitempty"`
	Transform  map[string]interface{} `json:"transform,omitempty" yaml:"transform,omitempty"`
	TrueNode   string                 `json:"true_node,omitempty" yaml:"true_node,omitempty"`
	FalseNode  string                 `json:"false_node,omitempty" yaml:"false_node,omitempty"`
}

// ValidationRule representa una regla de validación
type ValidationRule struct {
	Field    string      `json:"field" yaml:"field"`
	Type     string      `json:"type" yaml:"type"`
	Required bool        `json:"required,omitempty" yaml:"required,omitempty"`
	Pattern  string      `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Min      int         `json:"min,omitempty" yaml:"min,omitempty"`
	Max      int         `json:"max,omitempty" yaml:"max,omitempty"`
	MinValue interface{} `json:"min_value,omitempty" yaml:"min_value,omitempty"`
	MaxValue interface{} `json:"max_value,omitempty" yaml:"max_value,omitempty"`
	Message  string      `json:"message,omitempty" yaml:"message,omitempty"`
}

// WorkflowConfig representa la configuración completa del workflow
type WorkflowConfig struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string                 `json:"version,omitempty" yaml:"version,omitempty"`
	StartNode   string                 `json:"start_node" yaml:"start_node"`
	Nodes       []NodeConfig           `json:"nodes" yaml:"nodes"`
	Variables   map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty" yaml:"settings,omitempty"`
	Hooks       map[string]string      `json:"hooks,omitempty" yaml:"hooks,omitempty"`
}

// ExecutionContext contiene el contexto de ejecución del workflow
type ExecutionContext struct {
	WorkflowID  string                 `json:"workflow_id"`
	ExecutionID string                 `json:"execution_id"`
	NodeID      string                 `json:"node_id"`
	Data        interface{}            `json:"data"`
	Variables   map[string]interface{} `json:"variables"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"` // Campo agregado para los tests
	CurrentTime time.Time              `json:"current_time"`
	Error       error                  `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HookContext contiene el contexto para hooks
type HookContext struct {
	Stage         string                 `json:"stage"`
	WorkflowID    string                 `json:"workflow_id"`
	ExecutionID   string                 `json:"execution_id"`
	NodeID        string                 `json:"node_id,omitempty"`
	NodeType      NodeType               `json:"node_type,omitempty"`
	HookType      HookType               `json:"hook_type"`
	Data          interface{}            `json:"data"`
	Variables     map[string]interface{} `json:"variables"`
	ExecutionTime time.Duration          `json:"execution_time"`
	Error         error                  `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// WorkflowError representa un error específico del workflow
type WorkflowError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	NodeID     string                 `json:"node_id,omitempty"`
	NodeType   NodeType               `json:"node_type,omitempty"` // Campo agregado para los tests
	Cause      error                  `json:"cause,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Stack      string                 `json:"stack,omitempty"` // Alias para StackTrace
}

func (e *WorkflowError) Error() string {
	if e.NodeID != "" {
		return fmt.Sprintf("[%s] %s (node: %s)", e.Code, e.Message, e.NodeID)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *WorkflowError) Unwrap() error {
	return e.Cause
}

// NewWorkflowError crea un nuevo error de workflow
func NewWorkflowError(nodeID string, nodeType NodeType, message string, cause error) *WorkflowError {
	// Capturar stack trace
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackTrace := string(buf[:n])

	return &WorkflowError{
		Code:       nodeType.String(),
		Message:    message,
		NodeID:     nodeID,
		NodeType:   nodeType, // Asignar el NodeType
		Cause:      cause,
		Context:    make(map[string]interface{}),
		Timestamp:  time.Now(),
		StackTrace: stackTrace,
		Stack:      stackTrace, // Asignar Stack como alias de StackTrace
	}
}

// WithContext agrega contexto adicional al error
func (e *WorkflowError) WithContext(key string, value interface{}) *WorkflowError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Tipos para validaciones avanzadas
type AdvancedValidator struct {
	MaxDataSize       int           `json:"max_data_size,omitempty"`
	MaxValidationTime time.Duration `json:"max_validation_time,omitempty"`
}

type TypeConstraint struct {
	Type      string      `json:"type"`
	Required  bool        `json:"required"`
	Min       interface{} `json:"min,omitempty"`
	Max       interface{} `json:"max,omitempty"`
	MinLength int         `json:"min_length,omitempty"`
	MaxLength int         `json:"max_length,omitempty"`
	Pattern   string      `json:"pattern,omitempty"`
	Enum      []string    `json:"enum,omitempty"`
}

type AdvancedValidationResult struct {
	Valid   bool                   `json:"valid"`
	Errors  []ValidationError      `json:"errors"`
	Data    interface{}            `json:"data"`
	Summary map[string]interface{} `json:"summary"`
	Cycles  [][]string             `json:"cycles,omitempty"`
}

type GraphDefinition struct {
	Nodes []NodeConfig     `json:"nodes"`
	Edges []EdgeDefinition `json:"edges"`
}

type EdgeDefinition struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ComplexSchema struct {
	Properties map[string]PropertySchema `json:"properties"`
}

type PropertySchema struct {
	Type       string                    `json:"type"`
	Required   bool                      `json:"required"`
	Min        interface{}               `json:"min,omitempty"`
	Max        interface{}               `json:"max,omitempty"`
	MinLength  int                       `json:"min_length,omitempty"`
	MaxLength  int                       `json:"max_length,omitempty"`
	MinItems   int                       `json:"min_items,omitempty"`
	MaxItems   int                       `json:"max_items,omitempty"`
	Pattern    string                    `json:"pattern,omitempty"`
	Enum       []string                  `json:"enum,omitempty"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Items      *PropertySchema           `json:"items,omitempty"`
}

// ParseNodeType convierte un string a NodeType
func ParseNodeType(s string) (NodeType, error) {
	switch strings.ToLower(s) {
	case "task":
		return Task, nil
	case "subdag":
		return SubDag, nil
	case "conditional":
		return Conditional, nil
	case "foreach":
		return Foreach, nil
	case "branch":
		return Branch, nil
	case "parallel":
		return Parallel, nil
	case "http":
		return HTTP, nil
	case "delay":
		return Delay, nil
	case "transform":
		return Transform, nil
	case "validation":
		return Validation, nil
	case "merge":
		return Merge, nil
	case "split":
		return Split, nil
	case "filter":
		return Filter, nil
	default:
		return Task, fmt.Errorf("unknown node type: %s", s)
	}
}
