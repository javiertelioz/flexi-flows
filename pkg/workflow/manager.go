package workflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/storage"
)

type WorkflowManager struct {
	graph       *Graph
	stateStore  storage.StateStore
	hookManager HookManager
	logger      Logger
	metrics     MetricsCollector
	mu          sync.Mutex
	tasks       map[string]interface{}
	hooks       map[string]interface{}
}

func NewWorkflowManager() *WorkflowManager {
	return &WorkflowManager{
		graph:       &Graph{},
		hookManager: NewHookManager(),
		logger:      &DefaultLogger{},
		metrics:     NewDefaultMetricsCollector(),
		tasks:       make(map[string]interface{}),
		hooks:       make(map[string]interface{}),
	}
}

// SetLogger configura un logger personalizado
func (wm *WorkflowManager) SetLogger(logger Logger) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.logger = logger
}

// SetMetricsCollector configura un collector de métricas personalizado
func (wm *WorkflowManager) SetMetricsCollector(collector MetricsCollector) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.metrics = collector
}

// RegisterHookFunc registra una función como hook
func (wm *WorkflowManager) RegisterHookFunc(hookType HookType, fn interface{}) error {
	hook, err := NewFunctionHook(fn)
	if err != nil {
		return fmt.Errorf("failed to register hook function: %w", err)
	}
	wm.hookManager.RegisterHook(hookType, hook)
	return nil
}

// RegisterHook registra un hook personalizado
func (wm *WorkflowManager) RegisterHook(hookType HookType, hook Hook) {
	wm.hookManager.RegisterHook(hookType, hook)
}

// RegisterLoggingHook registra un hook de logging
func (wm *WorkflowManager) RegisterLoggingHook(hookType HookType, level LogLevel) {
	hook := &LoggingHook{
		Logger: wm.logger,
		Level:  level,
	}
	wm.hookManager.RegisterHook(hookType, hook)
}

// RegisterMetricsHook registra un hook de métricas
func (wm *WorkflowManager) RegisterMetricsHook(hookType HookType) {
	hook := &MetricsHook{
		Collector: wm.metrics,
	}
	wm.hookManager.RegisterHook(hookType, hook)
}

func (wm *WorkflowManager) AddNode(node NodeInterface) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.graph.Nodes = append(wm.graph.Nodes, node)
}

func (wm *WorkflowManager) AddEdge(edge *Edge) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.graph.Edges = append(wm.graph.Edges, edge)
}

// ExecuteWithContext ejecuta un workflow con contexto
func (wm *WorkflowManager) ExecuteWithContext(ctx context.Context, startNodeID string, initialData interface{}) (*ExecutionResult, error) {
	startTime := time.Now()

	// Crear contexto de ejecución
	executionID := fmt.Sprintf("exec_%d", time.Now().UnixNano())

	startNode := wm.findNodeByID(startNodeID)
	if startNode == nil {
		return nil, NewWorkflowError("", Task, "start node not found",
			errors.New("start node not found: "+startNodeID))
	}

	// Ejecutar hooks before execution
	hookCtx := &HookContext{
		NodeID:      startNode.GetID(),
		NodeType:    startNode.GetType(),
		HookType:    BeforeExecution,
		Data:        initialData,
		Metadata:    make(map[string]interface{}),
		ExecutionID: executionID,
		Timestamp:   startTime,
	}

	if err := wm.hookManager.ExecuteHooks(ctx, BeforeExecution, hookCtx); err != nil {
		return nil, fmt.Errorf("before execution hooks failed: %w", err)
	}

	// Ejecutar el workflow
	result, err := wm.ExecuteNodeWithContext(ctx, startNode, initialData)

	duration := time.Since(startTime)
	success := err == nil

	// Registrar métricas
	if wm.metrics != nil {
		wm.metrics.RecordNodeExecution(startNode.GetID(), startNode.GetType(), duration, success)
	}

	// Ejecutar hooks after execution
	hookCtx.HookType = AfterExecution
	hookCtx.Data = result
	hookCtx.Metadata["duration_ms"] = duration.Milliseconds()
	hookCtx.Metadata["success"] = success

	if success {
		hookCtx.HookType = OnSuccess
		if hookErr := wm.hookManager.ExecuteHooks(ctx, OnSuccess, hookCtx); hookErr != nil {
			wm.logger.Log(WARN, "success hooks failed", map[string]interface{}{
				"error":   hookErr.Error(),
				"node_id": startNode.GetID(),
			})
		}
	} else {
		hookCtx.HookType = OnError
		hookCtx.Metadata["error"] = err.Error()
		if hookErr := wm.hookManager.ExecuteHooks(ctx, OnError, hookCtx); hookErr != nil {
			wm.logger.Log(WARN, "error hooks failed", map[string]interface{}{
				"error":   hookErr.Error(),
				"node_id": startNode.GetID(),
			})
		}
	}

	// Siempre ejecutar hooks complete
	hookCtx.HookType = OnComplete
	if hookErr := wm.hookManager.ExecuteHooks(ctx, OnComplete, hookCtx); hookErr != nil {
		wm.logger.Log(WARN, "complete hooks failed", map[string]interface{}{
			"error":   hookErr.Error(),
			"node_id": startNode.GetID(),
		})
	}

	// Ejecutar hooks after execution
	hookCtx.HookType = AfterExecution
	if hookErr := wm.hookManager.ExecuteHooks(ctx, AfterExecution, hookCtx); hookErr != nil {
		wm.logger.Log(WARN, "after execution hooks failed", map[string]interface{}{
			"error":   hookErr.Error(),
			"node_id": startNode.GetID(),
		})
	}

	executionResult := &ExecutionResult{
		Data:      result,
		Metadata:  hookCtx.Metadata,
		Duration:  duration.Milliseconds(),
		Success:   success,
		Error:     err,
		NodeID:    startNode.GetID(),
		Timestamp: startTime.Unix(),
	}

	return executionResult, err
}

// ExecuteNodeWithContext ejecuta un nodo específico con contexto
func (wm *WorkflowManager) ExecuteNodeWithContext(ctx context.Context, node NodeInterface, data interface{}) (interface{}, error) {
	if node == nil {
		return nil, NewWorkflowError("", Task, "node is nil", errors.New("node is nil"))
	}

	nodeStartTime := time.Now()

	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(node.GetID(), node.GetType(), "context cancelled", ctx.Err())
	default:
	}

	// Crear contexto de hook para el nodo
	hookCtx := &HookContext{
		NodeID:      node.GetID(),
		NodeType:    node.GetType(),
		Data:        data,
		Metadata:    make(map[string]interface{}),
		ExecutionID: fmt.Sprintf("node_%s_%d", node.GetID(), time.Now().UnixNano()),
		Timestamp:   nodeStartTime,
	}

	// Cargar estado si existe
	if wm.stateStore != nil {
		state, err := wm.stateStore.LoadState(node.GetID())
		if err != nil {
			return nil, NewWorkflowError(node.GetID(), node.GetType(),
				"failed to load state", err)
		}
		if state != nil {
			data = state
			hookCtx.Data = data
		}
	}

	// Ejecutar hooks before
	hookCtx.HookType = BeforeExecution
	if err := wm.hookManager.ExecuteHooks(ctx, BeforeExecution, hookCtx); err != nil {
		return nil, fmt.Errorf("before hooks failed for node %s: %w", node.GetID(), err)
	}

	// Ejecutar el nodo
	result, err := node.Execute(ctx, wm, data)

	duration := time.Since(nodeStartTime)
	success := err == nil

	// Actualizar contexto de hook con resultado
	hookCtx.Data = result
	hookCtx.Metadata["duration_ms"] = duration.Milliseconds()
	hookCtx.Metadata["success"] = success

	if success {
		// Ejecutar hooks de éxito
		hookCtx.HookType = OnSuccess
		if hookErr := wm.hookManager.ExecuteHooks(ctx, OnSuccess, hookCtx); hookErr != nil {
			wm.logger.Log(WARN, "success hooks failed", map[string]interface{}{
				"error":   hookErr.Error(),
				"node_id": node.GetID(),
			})
		}
	} else {
		// Ejecutar hooks de error
		hookCtx.HookType = OnError
		hookCtx.Metadata["error"] = err.Error()
		if hookErr := wm.hookManager.ExecuteHooks(ctx, OnError, hookCtx); hookErr != nil {
			wm.logger.Log(WARN, "error hooks failed", map[string]interface{}{
				"error":   hookErr.Error(),
				"node_id": node.GetID(),
			})
		}
	}

	// Ejecutar hooks complete
	hookCtx.HookType = OnComplete
	if hookErr := wm.hookManager.ExecuteHooks(ctx, OnComplete, hookCtx); hookErr != nil {
		wm.logger.Log(WARN, "complete hooks failed", map[string]interface{}{
			"error":   err.Error(),
			"node_id": node.GetID(),
		})
	}

	// Ejecutar hooks after
	hookCtx.HookType = AfterExecution
	if hookErr := wm.hookManager.ExecuteHooks(ctx, AfterExecution, hookCtx); hookErr != nil {
		wm.logger.Log(WARN, "after hooks failed", map[string]interface{}{
			"error":   err.Error(),
			"node_id": node.GetID(),
		})
	}

	// Guardar estado si es necesario
	if wm.stateStore != nil && success {
		if err := wm.stateStore.SaveState(node.GetID(), result); err != nil {
			wm.logger.Log(WARN, "failed to save state", map[string]interface{}{
				"error":   err.Error(),
				"node_id": node.GetID(),
			})
		}
	}

	// Registrar métricas
	if wm.metrics != nil {
		wm.metrics.RecordNodeExecution(node.GetID(), node.GetType(), duration, success)
	}

	return result, err
}

// RegisterTask registra una tarea/función que puede ser usada por los nodos
func (wm *WorkflowManager) RegisterTask(name string, task interface{}) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.tasks[name] = task
}

// GetTask obtiene una tarea registrada por nombre
func (wm *WorkflowManager) GetTask(name string) (interface{}, bool) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	task, exists := wm.tasks[name]
	return task, exists
}

// LoadFromConfig carga un workflow desde configuración
func (wm *WorkflowManager) LoadFromConfig(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return wm.BuildFromConfig(cfg)
}

// BuildFromConfig construye un workflow desde una configuración
func (wm *WorkflowManager) BuildFromConfig(cfg *config.WorkflowConfig) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Limpiar el grafo actual
	wm.graph = &Graph{}

	// Crear nodos desde la configuración
	for _, nodeCfg := range cfg.Nodes {
		node, err := wm.createNodeFromConfig(&nodeCfg)
		if err != nil {
			return fmt.Errorf("failed to create node %s: %w", nodeCfg.ID, err)
		}
		wm.graph.Nodes = append(wm.graph.Nodes, node)
	}

	// Crear edges desde la configuración - usando los campos Next de cada nodo
	for _, nodeCfg := range cfg.Nodes {
		fromNode := wm.findNodeByID(nodeCfg.ID)
		if fromNode == nil {
			continue
		}

		// Crear edges para las conexiones Next
		for _, nextID := range nodeCfg.Next {
			toNode := wm.findNodeByID(nextID)
			if toNode != nil {
				edge := &Edge{
					From: fromNode,
					To:   toNode,
				}
				wm.graph.Edges = append(wm.graph.Edges, edge)
			}
		}

		// Crear edges para OnSuccess, OnError, etc.
		for _, successID := range nodeCfg.OnSuccess {
			toNode := wm.findNodeByID(successID)
			if toNode != nil {
				edge := &Edge{
					From: fromNode,
					To:   toNode,
				}
				wm.graph.Edges = append(wm.graph.Edges, edge)
			}
		}
	}

	return nil
}

// createNodeFromConfig crea un nodo desde una configuración
func (wm *WorkflowManager) createNodeFromConfig(nodeCfg *config.NodeConfig) (NodeInterface, error) {
	switch nodeCfg.Type {
	case "task":
		task, exists := wm.tasks[nodeCfg.Function]
		if !exists {
			return nil, fmt.Errorf("task %s not registered", nodeCfg.Function)
		}
		return &TaskNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Task,
			},
			Name: nodeCfg.Name,
			Task: task,
		}, nil

	case "conditional":
		conditionFunc, exists := wm.tasks[nodeCfg.Condition]
		if !exists {
			return nil, fmt.Errorf("condition task %s not registered", nodeCfg.Condition)
		}

		// Convertir la función a la firma esperada
		condition, ok := conditionFunc.(func(interface{}) bool)
		if !ok {
			return nil, fmt.Errorf("condition task %s does not have correct signature func(interface{}) bool", nodeCfg.Condition)
		}

		return &ConditionalNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Conditional,
			},
			Condition: condition,
			// TrueNext y FalseNext se configurarán después usando TruePath y FalsePath
		}, nil

	case "foreach":
		iterateFunc, exists := wm.tasks[nodeCfg.Iterator]
		if !exists {
			return nil, fmt.Errorf("iterator task %s not registered", nodeCfg.Iterator)
		}

		// Convertir la función a la firma esperada
		iterFunc, ok := iterateFunc.(func(interface{}) (interface{}, error))
		if !ok {
			return nil, fmt.Errorf("iterator task %s does not have correct signature func(interface{}) (interface{}, error)", nodeCfg.Iterator)
		}

		return &ForeachNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Foreach,
			},
			IterateFunc: iterFunc,
			// Collection se configurará desde CollectionArray si está disponible
			Collection: nodeCfg.CollectionArray,
		}, nil

	case "branch":
		return &BranchNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Branch,
			},
			// Branches se configurarán después usando ParallelTasks
		}, nil

	// ===== NUEVOS NODOS AVANZADOS =====

	case "http":
		if nodeCfg.URL == "" {
			return nil, fmt.Errorf("http node %s requires URL", nodeCfg.ID)
		}

		method := nodeCfg.Method
		if method == "" {
			method = "GET"
		}

		return &HTTPNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: HTTP,
			},
			URL:     nodeCfg.URL,
			Method:  method,
			Headers: nodeCfg.Headers,
			Body:    nodeCfg.Body,
			Timeout: time.Duration(30) * time.Second, // Default timeout
		}, nil

	case "delay":
		if nodeCfg.Duration == "" {
			return nil, fmt.Errorf("delay node %s requires duration", nodeCfg.ID)
		}

		duration, err := time.ParseDuration(nodeCfg.Duration)
		if err != nil {
			return nil, fmt.Errorf("invalid duration for delay node %s: %w", nodeCfg.ID, err)
		}

		return &DelayNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Delay,
			},
			Duration: duration,
		}, nil

	case "validation":
		if len(nodeCfg.Rules) == 0 {
			return nil, fmt.Errorf("validation node %s requires rules", nodeCfg.ID)
		}

		// Convertir reglas de configuración a ValidationRule
		var rules []ValidationRule
		for _, ruleConfig := range nodeCfg.Rules {
			rule := ValidationRule{
				Field:    ruleConfig.Field,
				Type:     ruleConfig.Type,
				Required: ruleConfig.Required,
				MinValue: ruleConfig.MinValue,
				MaxValue: ruleConfig.MaxValue,
				Pattern:  ruleConfig.Pattern,
				Message:  ruleConfig.Message,
			}
			rules = append(rules, rule)
		}

		return &ValidationNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Validation,
			},
			Rules:            rules,
			StopOnFirstError: false, // Configurar según necesidades
		}, nil

	case "transform":
		transformNode := &TransformNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Transform,
			},
			Mapping:      nodeCfg.Mapping,
			KeepOriginal: true, // Default behavior
		}

		// Si hay función de transformación personalizada registrada
		if nodeCfg.Transform != "" {
			if customFunc, exists := wm.tasks[nodeCfg.Transform]; exists {
				transformNode.CustomFunc = customFunc
			}
		}

		return transformNode, nil

	case "merge":
		strategy := MergeDeep // Default strategy
		if strategyStr, ok := nodeCfg.Mapping["strategy"].(string); ok {
			strategy = MergeStrategy(strategyStr)
		}

		return &MergeNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Merge,
			},
			Strategy:    strategy,
			IgnoreEmpty: true, // Default behavior
		}, nil

	case "split":
		splitNode := &SplitNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Split,
			},
			Strategy:     SplitByField, // Default strategy
			KeepOriginal: false,        // Default behavior
		}

		// Configurar estrategia desde mapping
		if mapping := nodeCfg.Mapping; mapping != nil {
			if strategyStr, ok := mapping["strategy"].(string); ok {
				splitNode.Strategy = SplitStrategy(strategyStr)
			}
			if pattern, ok := mapping["pattern"].(string); ok {
				splitNode.Pattern = pattern
			}
			if chunkSize, ok := mapping["chunk_size"].(float64); ok {
				splitNode.ChunkSize = int(chunkSize)
			}
			if fields, ok := mapping["fields"].([]interface{}); ok {
				for _, field := range fields {
					if fieldStr, ok := field.(string); ok {
						splitNode.Fields = append(splitNode.Fields, fieldStr)
					}
				}
			}
		}

		return splitNode, nil

	case "filter":
		filterNode := &FilterNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Filter,
			},
			Logic:       "and", // Default logic
			KeepMatched: true,  // Default behavior
		}

		// Configurar condiciones desde mapping
		if mapping := nodeCfg.Mapping; mapping != nil {
			if logic, ok := mapping["logic"].(string); ok {
				filterNode.Logic = logic
			}
			if keepMatched, ok := mapping["keep_matched"].(bool); ok {
				filterNode.KeepMatched = keepMatched
			}
			if conditions, ok := mapping["conditions"].([]interface{}); ok {
				for _, conditionData := range conditions {
					if condMap, ok := conditionData.(map[string]interface{}); ok {
						condition := FilterCondition{
							Field:    getStringFromMap(condMap, "field"),
							Operator: getStringFromMap(condMap, "operator"),
							Value:    condMap["value"],
						}
						if caseSensitive, exists := condMap["case_sensitive"]; exists {
							if cs, ok := caseSensitive.(bool); ok {
								condition.CaseSensitive = cs
							}
						}
						filterNode.Conditions = append(filterNode.Conditions, condition)
					}
				}
			}
		}

		return filterNode, nil

	case "subflow":
		subflowNode := &SubflowNode{
			Node: Node[interface{}]{
				ID:   nodeCfg.ID,
				Type: Subflow,
			},
			SubflowPath:  nodeCfg.SubflowPath,
			StartNodeID:  nodeCfg.SubflowID,
			IsolateState: true, // Default: isolate state
		}

		// Configurar parámetros desde mapping
		if nodeCfg.Mapping != nil {
			subflowNode.Parameters = nodeCfg.Mapping
		}

		return subflowNode, nil

	default:
		return nil, fmt.Errorf("unsupported node type: %s", nodeCfg.Type)
	}
}

// Helper function para extraer strings de maps
func getStringFromMap(m map[string]interface{}, key string) string {
	if value, exists := m[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// Execute ejecuta un workflow (método legacy para compatibilidad)
func (wm *WorkflowManager) Execute(startNodeID string, initialData interface{}) (interface{}, error) {
	ctx := context.Background()
	result, err := wm.ExecuteWithContext(ctx, startNodeID, initialData)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// findNodeByID busca un nodo por ID
func (wm *WorkflowManager) findNodeByID(nodeID string) NodeInterface {
	for _, node := range wm.graph.Nodes {
		if node.GetID() == nodeID {
			return node
		}
	}
	return nil
}

// GetGraph devuelve el grafo actual
func (wm *WorkflowManager) GetGraph() *Graph {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	return wm.graph
}

// SetStateStore configura el almacén de estado
func (wm *WorkflowManager) SetStateStore(store storage.StateStore) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.stateStore = store
}

// GetStateStore devuelve el almacén de estado actual
func (wm *WorkflowManager) GetStateStore() storage.StateStore {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	return wm.stateStore
}

// ValidateWorkflow valida que el workflow esté correctamente configurado
func (wm *WorkflowManager) ValidateWorkflow() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if len(wm.graph.Nodes) == 0 {
		return errors.New("workflow has no nodes")
	}

	// Validar que todos los nodos tengan IDs únicos
	nodeIDs := make(map[string]bool)
	for _, node := range wm.graph.Nodes {
		if nodeIDs[node.GetID()] {
			return fmt.Errorf("duplicate node ID: %s", node.GetID())
		}
		nodeIDs[node.GetID()] = true
	}

	// Validar edges
	for _, edge := range wm.graph.Edges {
		if !nodeIDs[edge.From.GetID()] {
			return fmt.Errorf("edge references non-existent node: %s", edge.From.GetID())
		}
		if !nodeIDs[edge.To.GetID()] {
			return fmt.Errorf("edge references non-existent node: %s", edge.To.GetID())
		}
	}

	return nil
}
