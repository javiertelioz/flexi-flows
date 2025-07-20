package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// WorkflowConfig representa la configuración completa del workflow
type WorkflowConfig struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string                 `json:"version,omitempty" yaml:"version,omitempty"`
	StartNode   string                 `json:"start_node" yaml:"start_node"`
	Variables   map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	Nodes       []NodeConfig           `json:"nodes" yaml:"nodes"`
	Hooks       map[string]HookConfig  `json:"hooks,omitempty" yaml:"hooks,omitempty"`
	Settings    WorkflowSettings       `json:"settings,omitempty" yaml:"settings,omitempty"`
	// Para compatibilidad con el sistema anterior
	Edges []EdgeConfig `json:"edges,omitempty" yaml:"edges,omitempty"`
}

// WorkflowSettings configuraciones globales del workflow
type WorkflowSettings struct {
	Timeout        string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	MaxRetries     int    `json:"max_retries,omitempty" yaml:"max_retries,omitempty"`
	ParallelLimit  int    `json:"parallel_limit,omitempty" yaml:"parallel_limit,omitempty"`
	EnableDebug    bool   `json:"enable_debug,omitempty" yaml:"enable_debug,omitempty"`
	EnableMetrics  bool   `json:"enable_metrics,omitempty" yaml:"enable_metrics,omitempty"`
	EnableAuditLog bool   `json:"enable_audit_log,omitempty" yaml:"enable_audit_log,omitempty"`
}

// EdgeConfig para compatibilidad con el sistema anterior
type EdgeConfig struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

// ConfigParser maneja la lectura y procesamiento de configuraciones
type ConfigParser struct {
	templateEngine *TemplateEngine
}

// NewConfigParser crea un nuevo parser de configuraciones
func NewConfigParser() *ConfigParser {
	return &ConfigParser{
		templateEngine: NewTemplateEngine(),
	}
}

// ParseFromFile carga y parsea una configuración desde un archivo
func (cp *ConfigParser) ParseFromFile(filePath string) (*WorkflowConfig, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return cp.ParseFromJSON(data)
	case ".yaml", ".yml":
		return cp.ParseFromYAML(data)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// ParseFromJSON parsea una configuración desde JSON
func (cp *ConfigParser) ParseFromJSON(data []byte) (*WorkflowConfig, error) {
	var config WorkflowConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return cp.processConfig(&config)
}

// ParseFromYAML parsea una configuración desde YAML
func (cp *ConfigParser) ParseFromYAML(data []byte) (*WorkflowConfig, error) {
	var config WorkflowConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return cp.processConfig(&config)
}

// processConfig procesa y valida una configuración cargada
func (cp *ConfigParser) processConfig(config *WorkflowConfig) (*WorkflowConfig, error) {
	// Aplicar templating con variables
	if err := cp.applyTemplating(config); err != nil {
		return nil, fmt.Errorf("failed to apply templating: %w", err)
	}

	// Validar configuración
	if err := cp.validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Convertir navegación simplificada a edges si es necesario
	if err := cp.expandNavigation(config); err != nil {
		return nil, fmt.Errorf("failed to expand navigation: %w", err)
	}

	return config, nil
}

// applyTemplating aplica el sistema de templates con variables preservando tipos
func (cp *ConfigParser) applyTemplating(config *WorkflowConfig) error {
	if len(config.Variables) == 0 {
		return nil
	}

	// En lugar de convertir toda la configuración a string y de vuelta,
	// procesamos campo por campo para preservar tipos
	for i := range config.Nodes {
		node := &config.Nodes[i]

		// Procesar campos de string
		if node.URL != "" {
			processed, _ := cp.templateEngine.Process(node.URL, config.Variables)
			node.URL = processed
		}
		if node.Timeout != "" {
			processed, _ := cp.templateEngine.Process(node.Timeout, config.Variables)
			node.Timeout = processed
		}
		if node.Function != "" {
			processed, _ := cp.templateEngine.Process(node.Function, config.Variables)
			node.Function = processed
		}
		if node.Collection != "" {
			processed, _ := cp.templateEngine.Process(node.Collection, config.Variables)
			node.Collection = processed
		}
		if node.Condition != "" {
			processed, _ := cp.templateEngine.Process(node.Condition, config.Variables)
			node.Condition = processed
		}
		if node.Duration != "" {
			processed, _ := cp.templateEngine.Process(node.Duration, config.Variables)
			node.Duration = processed
		}

		// Procesar headers si existen
		for key, value := range node.Headers {
			processed, _ := cp.templateEngine.Process(value, config.Variables)
			node.Headers[key] = processed
		}

		// Para campos que pueden ser variables de número (como retry_count),
		// los procesamos especialmente
		// Por ahora, mantenemos el enfoque simple para esta prueba
	}

	return nil
}

// validateConfig valida la configuración cargada
func (cp *ConfigParser) validateConfig(config *WorkflowConfig) error {
	if config.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if config.StartNode == "" {
		return fmt.Errorf("start_node is required")
	}

	if len(config.Nodes) == 0 {
		return fmt.Errorf("at least one node is required")
	}

	// Validar que el nodo inicial existe
	startNodeExists := false
	nodeIDs := make(map[string]bool)

	for _, node := range config.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}

		if node.Type == "" {
			return fmt.Errorf("node type is required for node %s", node.ID)
		}

		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true

		if node.ID == config.StartNode {
			startNodeExists = true
		}

		// Validar nodos específicos
		if err := cp.validateNodeConfig(&node); err != nil {
			return fmt.Errorf("invalid config for node %s: %w", node.ID, err)
		}
	}

	if !startNodeExists {
		return fmt.Errorf("start node %s not found in nodes", config.StartNode)
	}

	// Validar referencias de navegación
	if err := cp.validateNavigation(config, nodeIDs); err != nil {
		return fmt.Errorf("navigation validation failed: %w", err)
	}

	return nil
}

// validateNodeConfig valida la configuración específica de un nodo
func (cp *ConfigParser) validateNodeConfig(node *NodeConfig) error {
	switch node.Type {
	case "task":
		if node.Function == "" && node.TaskFunc == "" {
			return fmt.Errorf("task nodes require a function")
		}
	case "http":
		if node.URL == "" {
			return fmt.Errorf("HTTP nodes require a URL")
		}
		if node.Method == "" {
			node.Method = "GET" // Default
		}
	case "delay":
		if node.Duration == "" {
			return fmt.Errorf("delay nodes require a duration")
		}
		// Validar formato de duración
		if _, err := time.ParseDuration(node.Duration); err != nil {
			return fmt.Errorf("invalid duration format: %s", node.Duration)
		}
	case "conditional":
		if node.Condition == "" {
			return fmt.Errorf("conditional nodes require a condition")
		}
	case "loop":
		if node.Collection == "" {
			return fmt.Errorf("loop nodes require a collection")
		}
		if node.Iterator == "" {
			node.Iterator = "item" // Default
		}
	case "validation":
		if len(node.Rules) == 0 {
			return fmt.Errorf("validation nodes require at least one rule")
		}
	case "parallel":
		if len(node.ParallelTasks) == 0 {
			return fmt.Errorf("parallel nodes require parallel tasks")
		}
	}

	return nil
}

// validateNavigation valida las referencias de navegación
func (cp *ConfigParser) validateNavigation(config *WorkflowConfig, nodeIDs map[string]bool) error {
	for _, node := range config.Nodes {
		// Validar Next
		for _, nextID := range node.Next {
			if !nodeIDs[nextID] {
				return fmt.Errorf("node %s references non-existent node %s in next", node.ID, nextID)
			}
		}

		// Validar OnSuccess
		for _, successID := range node.OnSuccess {
			if !nodeIDs[successID] {
				return fmt.Errorf("node %s references non-existent node %s in on_success", node.ID, successID)
			}
		}

		// Validar OnError
		for _, errorID := range node.OnError {
			if !nodeIDs[errorID] {
				return fmt.Errorf("node %s references non-existent node %s in on_error", node.ID, errorID)
			}
		}

		// Validar TruePath y FalsePath para condicionales
		for _, trueID := range node.TruePath {
			if !nodeIDs[trueID] {
				return fmt.Errorf("node %s references non-existent node %s in true_path", node.ID, trueID)
			}
		}

		for _, falseID := range node.FalsePath {
			if !nodeIDs[falseID] {
				return fmt.Errorf("node %s references non-existent node %s in false_path", node.ID, falseID)
			}
		}

		// Validar ParallelTasks
		for _, parallelID := range node.ParallelTasks {
			if !nodeIDs[parallelID] {
				return fmt.Errorf("node %s references non-existent node %s in parallel_tasks", node.ID, parallelID)
			}
		}
	}

	return nil
}

// expandNavigation convierte la navegación simplificada en edges para compatibilidad
func (cp *ConfigParser) expandNavigation(config *WorkflowConfig) error {
	// Si ya tiene edges definidos, no expandir
	if len(config.Edges) > 0 {
		return nil
	}

	var edges []EdgeConfig

	for _, node := range config.Nodes {
		// Next edges
		for _, next := range node.Next {
			edges = append(edges, EdgeConfig{From: node.ID, To: next})
		}

		// Success edges
		for _, success := range node.OnSuccess {
			edges = append(edges, EdgeConfig{From: node.ID, To: success})
		}

		// Error edges
		for _, errorNode := range node.OnError {
			edges = append(edges, EdgeConfig{From: node.ID, To: errorNode})
		}

		// True path edges
		for _, trueNode := range node.TruePath {
			edges = append(edges, EdgeConfig{From: node.ID, To: trueNode})
		}

		// False path edges
		for _, falseNode := range node.FalsePath {
			edges = append(edges, EdgeConfig{From: node.ID, To: falseNode})
		}

		// Parallel task edges
		for _, parallelNode := range node.ParallelTasks {
			edges = append(edges, EdgeConfig{From: node.ID, To: parallelNode})
		}
	}

	config.Edges = edges
	return nil
}

// TemplateEngine maneja el procesamiento de templates
type TemplateEngine struct {
	variableRegex *regexp.Regexp
}

// NewTemplateEngine crea un nuevo motor de templates
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		variableRegex: regexp.MustCompile(`\$\{([^}]+)\}`),
	}
}

// Process procesa un texto aplicando las variables
func (te *TemplateEngine) Process(text string, variables map[string]interface{}) (string, error) {
	result := te.variableRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Extraer el nombre de la variable (sin ${ y })
		varName := strings.Trim(match, "${}")

		if value, exists := variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}

		// Si la variable no existe, dejar el placeholder
		return match
	})

	return result, nil
}
