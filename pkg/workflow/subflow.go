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
	SubflowPath    string                 // Ruta al archivo de configuración del sub-workflow
	SubflowConfig  *config.WorkflowConfig // Configuración del sub-workflow (si se proporciona directamente)
	SubflowManager *WorkflowManager       // Manager para el sub-workflow
	StartNodeID    string                 // ID del nodo inicial en el sub-workflow
	Parameters     map[string]interface{} // Parámetros a pasar al sub-workflow
	Timeout        int64                  // Timeout en segundos para el sub-workflow
	IsolateState   bool                   // Si aislar el estado del sub-workflow
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

	// Crear manager para el subflow si no existe
	if sn.SubflowManager == nil {
		sn.SubflowManager = NewWorkflowManager()

		// Cargar configuración del subworkflow si se especifica una ruta
		if sn.SubflowPath != "" {
			if err := sn.loadSubflowConfig(sn.SubflowPath); err != nil {
				return nil, NewWorkflowError(sn.ID, sn.Type, "failed to load subflow configuration", err)
			}
		}
	}

	// Validar que hay un nodo de inicio
	if sn.StartNodeID == "" {
		return nil, NewWorkflowError(sn.ID, sn.Type, "no start node specified for subflow", fmt.Errorf("empty start node ID"))
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

	// Ejecutar el subworkflow
	result, err := subWm.ExecuteWithContext(subCtx, sn.StartNodeID, inputData)
	if err != nil {
		return nil, NewWorkflowError(sn.ID, sn.Type, "subflow execution failed", err)
	}

	duration := getCurrentTimeMillis() - startTime

	// Preparar resultado
	subflowResult := &SubflowResult{
		Duration:    duration,
		SubflowName: "subflow", // Usar un nombre por defecto
		StartNode:   sn.StartNodeID,
		Metadata: map[string]interface{}{
			"subflow_path":    sn.SubflowPath,
			"isolated_state":  sn.IsolateState,
			"timeout_seconds": sn.Timeout,
			"parameters":      sn.Parameters,
		},
	}

	subflowResult.Success = true
	if result != nil {
		subflowResult.Data = result.Data
		subflowResult.Metadata["execution_result"] = result
	}

	return subflowResult, nil
}

// loadSubflowConfig carga la configuración del subworkflow desde un archivo
func (sn *SubflowNode) loadSubflowConfig(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load subflow config from %s: %w", configPath, err)
	}

	sn.SubflowConfig = cfg

	// Construir el subworkflow desde la configuración
	if sn.SubflowManager == nil {
		sn.SubflowManager = NewWorkflowManager()
	}

	return sn.SubflowManager.BuildFromConfig(cfg)
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
		// No propagar errores de cancelación de contexto usando el campo Code en lugar de Type
		if workflowErr.Code == "context cancelled" {
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
