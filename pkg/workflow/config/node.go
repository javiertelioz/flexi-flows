package config

// NodeConfig representa la configuración de un nodo individual
// Combina la configuración anterior con las nuevas características
type NodeConfig struct {
	ID          string `json:"id" yaml:"id"`
	Type        string `json:"type" yaml:"type"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Function    string `json:"function,omitempty" yaml:"function,omitempty"`

	// Navegación simplificada
	Next       []string `json:"next,omitempty" yaml:"next,omitempty"`
	OnSuccess  []string `json:"on_success,omitempty" yaml:"on_success,omitempty"`
	OnError    []string `json:"on_error,omitempty" yaml:"on_error,omitempty"`
	OnComplete []string `json:"on_complete,omitempty" yaml:"on_complete,omitempty"`

	// Para nodos condicionales
	Condition string   `json:"condition,omitempty" yaml:"condition,omitempty"`
	TruePath  []string `json:"true_path,omitempty" yaml:"true_path,omitempty"`
	FalsePath []string `json:"false_path,omitempty" yaml:"false_path,omitempty"`

	// Para nodos de loop
	Iterator   string `json:"iterator,omitempty" yaml:"iterator,omitempty"`
	Collection string `json:"collection,omitempty" yaml:"collection,omitempty"`
	// Para compatibilidad con el sistema anterior - colección como array
	CollectionArray []interface{} `json:"collectionArray,omitempty" yaml:"collectionArray,omitempty"`

	// Para nodos paralelos
	ParallelTasks []string `json:"parallel_tasks,omitempty" yaml:"parallel_tasks,omitempty"`

	// Para nodos HTTP
	URL     string                 `json:"url,omitempty" yaml:"url,omitempty"`
	Method  string                 `json:"method,omitempty" yaml:"method,omitempty"`
	Headers map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body    map[string]interface{} `json:"body,omitempty" yaml:"body,omitempty"`

	// Para nodos de validación
	Rules []ValidationRuleConfig `json:"rules,omitempty" yaml:"rules,omitempty"`

	// Para nodos de delay
	Duration string `json:"duration,omitempty" yaml:"duration,omitempty"`

	// Para nodos de transformación
	Transform string                 `json:"transform,omitempty" yaml:"transform,omitempty"`
	Mapping   map[string]interface{} `json:"mapping,omitempty" yaml:"mapping,omitempty"`

	// Para subflows
	SubflowPath string `json:"subflow_path,omitempty" yaml:"subflow_path,omitempty"`
	SubflowID   string `json:"subflow_id,omitempty" yaml:"subflow_id,omitempty"`

	// Configuración de hooks
	Hooks map[string][]HookConfig `json:"hooks,omitempty" yaml:"hooks,omitempty"`

	// Para configuración de timeouts y advanced features
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// TaskFunc para compatibilidad con nodos de tarea
	TaskFunc interface{} `json:"task_func,omitempty" yaml:"task_func,omitempty"`
}

// ValidationRuleConfig representa la configuración de una regla de validación
type ValidationRuleConfig struct {
	Field    string      `json:"field" yaml:"field"`
	Type     string      `json:"type" yaml:"type"`
	Required bool        `json:"required,omitempty" yaml:"required,omitempty"`
	MinValue interface{} `json:"min,omitempty" yaml:"min,omitempty"`
	MaxValue interface{} `json:"max,omitempty" yaml:"max,omitempty"`
	Pattern  string      `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Message  string      `json:"message,omitempty" yaml:"message,omitempty"`
}

// HookConfig representa la configuración de un hook
type HookConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Function string                 `json:"function" yaml:"function"`
	Enabled  *bool                  `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
}
