package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

// SubflowNode representa un nodo que ejecuta un sub-workflow
type SubflowNode struct {
	Node[interface{}]
	SubflowPath   string                 // Ruta al archivo de configuración del sub-workflow
	SubflowConfig *config.WorkflowConfig // Configuración del sub-workflow (si se proporciona directamente)
	StartNodeID   string                 // ID del nodo inicial en el sub-workflow
	Parameters    map[string]interface{} // Parámetros a pasar al sub-workflow
	Timeout       int64                  // Timeout en segundos para el sub-workflow
	IsolateState  bool                   // Si aislar el estado del sub-workflow
}

// SubflowResult contiene el resultado de la ejecución del sub-workflow
type SubflowResult struct {
	Data        interface{}            `json:"data"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Duration    int64                  `json:"duration_ms"`
	SubflowName string                 `json:"subflow_name"`
	StartNode   string                 `json:"start_node"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Execute ejecuta el sub-workflow y retorna el resultado
func (sn *SubflowNode) Execute(ctx context.Context, wm *WorkflowManager, data interface{}) (interface{}, error) {
	// Verificar si el contexto ha sido cancelado
	select {
	case <-ctx.Done():
		return nil, NewWorkflowError(sn.ID, sn.Type, "context cancelled before subflow execution", ctx.Err())
	default:
	}

	startTime := getCurrentTimeMillis()

	// Crear un nuevo WorkflowManager para el sub-workflow si se requiere aislamiento
	var subWm *WorkflowManager
	if sn.IsolateState {
		subWm = NewWorkflowManager()
		// Copiar las tareas registradas del manager principal
		sn.copyTasks(wm, subWm)
	} else {
		subWm = wm
	}

	// Cargar configuración del sub-workflow
	var cfg *config.WorkflowConfig
	var err error

	if sn.SubflowConfig != nil {
		cfg = sn.SubflowConfig
	} else if sn.SubflowPath != "" {
		cfg, err = config.LoadConfig(sn.SubflowPath)
		if err != nil {
			return nil, NewWorkflowError(sn.ID, sn.Type,
				fmt.Sprintf("failed to load subflow config from %s", sn.SubflowPath), err)
		}
	} else {
		return nil, NewWorkflowError(sn.ID, sn.Type,
			"no subflow configuration provided", fmt.Errorf("either SubflowPath or SubflowConfig must be specified"))
	}

	// Construir el sub-workflow
	if err := subWm.BuildFromConfig(cfg); err != nil {
		return nil, NewWorkflowError(sn.ID, sn.Type, "failed to build subflow", err)
	}

	// Determinar nodo de inicio
	startNodeID := sn.StartNodeID
	if startNodeID == "" {
		startNodeID = cfg.StartNode
	}
	if startNodeID == "" {
		return nil, NewWorkflowError(sn.ID, sn.Type,
			"no start node specified for subflow", fmt.Errorf("StartNodeID or config.StartNode must be specified"))
	}

	// Preparar datos de entrada
	inputData := data
	if sn.Parameters != nil {
		// Fusionar parámetros con datos de entrada
		inputData = sn.mergeDataWithParameters(data, sn.Parameters)
	}

	// Crear contexto con timeout si se especifica
	subCtx := ctx
	if sn.Timeout > 0 {
		var cancel context.CancelFunc
		subCtx, cancel = context.WithTimeout(ctx, getDurationFromSeconds(sn.Timeout))
		defer cancel()
	}

	// Ejecutar el sub-workflow
	result, err := subWm.ExecuteWithContext(subCtx, startNodeID, inputData)

	duration := getCurrentTimeMillis() - startTime

	// Preparar resultado
	subflowResult := &SubflowResult{
		Duration:    duration,
		SubflowName: cfg.Name,
		StartNode:   startNodeID,
		Metadata: map[string]interface{}{
			"subflow_path":    sn.SubflowPath,
			"isolated_state":  sn.IsolateState,
			"timeout_seconds": sn.Timeout,
			"parameters":      sn.Parameters,
		},
	}

	if err != nil {
		subflowResult.Success = false
		subflowResult.Error = err.Error()

		// Decidir si propagar el error o solo reportarlo
		if sn.shouldPropagateError(err) {
			return subflowResult, NewWorkflowError(sn.ID, sn.Type,
				fmt.Sprintf("subflow execution failed: %s", err.Error()), err)
		}
	} else {
		subflowResult.Success = true
		if result != nil {
			subflowResult.Data = result.Data
			subflowResult.Metadata["execution_result"] = result
		}
	}

	return subflowResult, nil
}

// copyTasks copia las tareas registradas de un manager a otro
func (sn *SubflowNode) copyTasks(source, target *WorkflowManager) {
	// Acceder a las tareas del source manager
	// Nota: Esto requiere que el campo tasks sea público o que haya un método getter
	source.mu.Lock()
	defer source.mu.Unlock()

	target.mu.Lock()
	defer target.mu.Unlock()

	// Copiar todas las tareas
	for name, task := range source.tasks {
		target.tasks[name] = task
	}
}

// mergeDataWithParameters fusiona los datos de entrada con los parámetros
func (sn *SubflowNode) mergeDataWithParameters(data interface{}, parameters map[string]interface{}) interface{} {
	// Si los datos son un map, fusionar con parámetros
	if dataMap, ok := data.(map[string]interface{}); ok {
		merged := make(map[string]interface{})

		// Copiar datos originales
		for k, v := range dataMap {
			merged[k] = v
		}

		// Agregar parámetros (sobrescribir si existen)
		for k, v := range parameters {
			merged[k] = v
		}

		return merged
	}

	// Si los datos no son un map, crear uno nuevo con datos y parámetros
	result := make(map[string]interface{})
	result["data"] = data

	for k, v := range parameters {
		result[k] = v
	}

	return result
}

// shouldPropagateError determina si un error debe ser propagado
func (sn *SubflowNode) shouldPropagateError(err error) bool {
	// Por defecto, propagar todos los errores
	// Se pueden agregar reglas más sofisticadas aquí
	if workflowErr, ok := err.(*WorkflowError); ok {
		// No propagar errores de cancelación de contexto
		if workflowErr.Type == "context cancelled" {
			return false
		}
	}
	return true
}

// getCurrentTimeMillis obtiene el tiempo actual en milisegundos
func getCurrentTimeMillis() int64 {
	return int64(time.Now().UnixNano() / int64(time.Millisecond))
}

// getDurationFromSeconds convierte segundos a time.Duration
func getDurationFromSeconds(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}
