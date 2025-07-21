package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

func main() {
	// Crear el manager de workflows con el nuevo sistema de hooks
	wm := workflow.NewWorkflowManager()

	// Configurar logging personalizado
	logger := &CustomLogger{}
	wm.SetLogger(logger)

	// Configurar métricas personalizadas
	metrics := &CustomMetricsCollector{}
	wm.SetMetricsCollector(metrics)

	// Registrar hooks globales
	registerGlobalHooks(wm)

	// Registrar funciones de tareas
	registerTaskFunctions(wm)

	// Crear y ejecutar el workflow
	executeWorkflow(wm)

	// Mostrar métricas finales
	metrics.PrintSummary()
}

// CustomLogger implementa logging estructurado
type CustomLogger struct{}

func (cl *CustomLogger) Log(level workflow.LogLevel, message string, fields map[string]interface{}) {
	levelStr := map[workflow.LogLevel]string{
		workflow.DEBUG: "DEBUG",
		workflow.INFO:  "INFO",
		workflow.WARN:  "WARN",
		workflow.ERROR: "ERROR",
	}[level]

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s - %s", timestamp, levelStr, message)

	if len(fields) > 0 {
		fmt.Printf(" | Fields: %+v", fields)
	}
	fmt.Println()
}

// CustomMetricsCollector recopila métricas personalizadas
type CustomMetricsCollector struct {
	totalNodes     int
	successfulRuns int
	failedRuns     int
	totalDuration  time.Duration
	hookMetrics    map[string]int
}

func (cmc *CustomMetricsCollector) RecordHookExecution(hookType workflow.HookType, nodeType workflow.NodeType, duration time.Duration) {
	if cmc.hookMetrics == nil {
		cmc.hookMetrics = make(map[string]int)
	}
	cmc.hookMetrics[string(hookType)]++
}

func (cmc *CustomMetricsCollector) RecordNodeExecution(nodeID string, nodeType workflow.NodeType, duration time.Duration, success bool) {
	cmc.totalNodes++
	cmc.totalDuration += duration

	if success {
		cmc.successfulRuns++
	} else {
		cmc.failedRuns++
	}
}

func (cmc *CustomMetricsCollector) IncrementCounter(name string, tags map[string]string) {
	if cmc.hookMetrics == nil {
		cmc.hookMetrics = make(map[string]int)
	}
	cmc.hookMetrics[name]++
}

func (cmc *CustomMetricsCollector) PrintSummary() {
	fmt.Println("\n=== WORKFLOW EXECUTION SUMMARY ===")
	fmt.Printf("Total Nodes Executed: %d\n", cmc.totalNodes)
	fmt.Printf("Successful Runs: %d\n", cmc.successfulRuns)
	fmt.Printf("Failed Runs: %d\n", cmc.failedRuns)
	fmt.Printf("Total Duration: %v\n", cmc.totalDuration)
	fmt.Printf("Average Duration per Node: %v\n", cmc.totalDuration/time.Duration(max(1, cmc.totalNodes)))

	fmt.Println("\nHook Metrics:")
	for metric, count := range cmc.hookMetrics {
		fmt.Printf("  %s: %d\n", metric, count)
	}
	fmt.Println("=====================================")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// registerGlobalHooks configura hooks globales para el workflow
func registerGlobalHooks(wm *workflow.WorkflowManager) {
	// Hook de auditoría que se ejecuta antes de cada nodo
	auditHook := func(ctx context.Context, hookCtx *workflow.HookContext) error {
		fmt.Printf("🔍 AUDIT: Starting execution of node '%s' (type: %s)\n",
			hookCtx.NodeID, hookCtx.NodeType)
		return nil
	}

	err := wm.RegisterHookFunc(workflow.BeforeExecution, auditHook)
	if err != nil {
		log.Printf("Error registering audit hook: %v", err)
	}

	// Hook de logging para errores
	errorLogHook := func(ctx context.Context, hookCtx *workflow.HookContext) error {
		if errorMsg, exists := hookCtx.Metadata["error"]; exists {
			fmt.Printf("❌ ERROR in node '%s': %v\n", hookCtx.NodeID, errorMsg)
		}
		return nil
	}

	err = wm.RegisterHookFunc(workflow.OnError, errorLogHook)
	if err != nil {
		log.Printf("Error registering error log hook: %v", err)
	}

	// Hook de éxito con información de timing
	successHook := func(ctx context.Context, hookCtx *workflow.HookContext) error {
		duration := hookCtx.Metadata["duration_ms"]
		fmt.Printf("✅ SUCCESS: Node '%s' completed in %vms\n",
			hookCtx.NodeID, duration)
		return nil
	}

	err = wm.RegisterHookFunc(workflow.OnSuccess, successHook)
	if err != nil {
		log.Printf("Error registering success hook: %v", err)
	}

	// Hook condicional que solo se ejecuta para nodos específicos
	conditionalHook := &workflow.ConditionalHook{
		Condition: func(ctx context.Context, hookCtx *workflow.HookContext) bool {
			return hookCtx.NodeType == workflow.Task
		},
		Hook: &workflow.FunctionHook{
			Func: func() error {
				fmt.Println("🎯 Conditional hook executed for Task node")
				return nil
			},
		},
	}
	wm.RegisterHook(workflow.AfterExecution, conditionalHook)

	// Hook en cadena para finalización
	hook1, _ := workflow.NewFunctionHook(func() error {
		fmt.Println("🔗 Chain Hook 1: Cleanup resources")
		return nil
	})

	hook2, _ := workflow.NewFunctionHook(func() error {
		fmt.Println("🔗 Chain Hook 2: Send notifications")
		return nil
	})

	chainHook := &workflow.ChainHook{
		Hooks: []workflow.Hook{hook1, hook2},
	}
	wm.RegisterHook(workflow.OnComplete, chainHook)
}

// registerTaskFunctions registra las funciones de tarea del workflow
func registerTaskFunctions(wm *workflow.WorkflowManager) {
	// Tarea de procesamiento de datos
	wm.RegisterTask("ProcessData", func(data map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🔄 Processing data...")
		time.Sleep(100 * time.Millisecond) // Simular trabajo

		data["processed"] = true
		data["timestamp"] = time.Now().Unix()
		data["processing_duration"] = "100ms"

		return data, nil
	})

	// Tarea de validación
	wm.RegisterTask("ValidateData", func(data map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("✔️ Validating data...")

		if _, exists := data["processed"]; !exists {
			return nil, errors.New("data has not been processed")
		}

		data["validated"] = true
		return data, nil
	})

	// Tarea de transformación
	wm.RegisterTask("TransformData", func(data map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🔄 Transforming data...")

		// Simular transformación compleja
		time.Sleep(200 * time.Millisecond)

		data["transformed"] = true
		data["format"] = "v2"

		return data, nil
	})

	// Tarea de finalización
	wm.RegisterTask("FinalizeData", func(data map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🎯 Finalizing data...")

		data["finalized"] = true
		data["completion_time"] = time.Now().Format("2006-01-02 15:04:05")

		return data, nil
	})

	// Tarea que puede fallar (para demostrar manejo de errores)
	wm.RegisterTask("RiskyOperation", func(data map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("⚠️ Executing risky operation...")

		// Simular una operación que puede fallar
		if time.Now().UnixNano()%2 == 0 {
			return nil, errors.New("risky operation failed randomly")
		}

		data["risky_completed"] = true
		return data, nil
	})
}

// executeWorkflow crea y ejecuta un workflow de demostración
func executeWorkflow(wm *workflow.WorkflowManager) {
	// Crear nodos del workflow
	processNode := &workflow.Node[interface{}]{
		ID:   "process_data",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			return wm.tasks["ProcessData"].(func(map[string]interface{}) (map[string]interface{}, error))(data.(map[string]interface{}))
		},
	}

	validateNode := &workflow.Node[interface{}]{
		ID:   "validate_data",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			return wm.tasks["ValidateData"].(func(map[string]interface{}) (map[string]interface{}, error))(data.(map[string]interface{}))
		},
	}

	transformNode := &workflow.Node[interface{}]{
		ID:   "transform_data",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			return wm.tasks["TransformData"].(func(map[string]interface{}) (map[string]interface{}, error))(data.(map[string]interface{}))
		},
	}

	finalizeNode := &workflow.Node[interface{}]{
		ID:   "finalize_data",
		Type: workflow.Task,
		TaskFunc: func(data interface{}) (interface{}, error) {
			return wm.tasks["FinalizeData"].(func(map[string]interface{}) (map[string]interface{}, error))(data.(map[string]interface{}))
		},
	}

	// Configurar el flujo
	processNode.Next = []workflow.NodeInterface{validateNode}
	validateNode.Next = []workflow.NodeInterface{transformNode}
	transformNode.Next = []workflow.NodeInterface{finalizeNode}

	// Agregar nodos al workflow
	wm.AddNode(processNode)
	wm.AddNode(validateNode)
	wm.AddNode(transformNode)
	wm.AddNode(finalizeNode)

	// Datos iniciales
	initialData := map[string]interface{}{
		"input":  "raw data",
		"source": "example system",
		"id":     "workflow-demo-001",
	}

	fmt.Println("\n🚀 Starting Advanced Workflow with Context and Hooks")
	fmt.Println("===================================================")

	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ejecutar el workflow
	result, err := wm.ExecuteWithContext(ctx, "process_data", initialData)

	if err != nil {
		fmt.Printf("\n❌ Workflow execution failed: %v\n", err)

		// Si es un WorkflowError, mostrar información detallada
		if workflowErr, ok := err.(*workflow.WorkflowError); ok {
			fmt.Printf("Node ID: %s\n", workflowErr.NodeID)
			fmt.Printf("Node Type: %s\n", workflowErr.NodeType)
			fmt.Printf("Timestamp: %v\n", workflowErr.Timestamp)
			if len(workflowErr.Context) > 0 {
				fmt.Printf("Context: %+v\n", workflowErr.Context)
			}
		}
	} else {
		fmt.Printf("\n✅ Workflow completed successfully!\n")
		fmt.Printf("Execution Duration: %dms\n", result.Duration)
		fmt.Printf("Success: %t\n", result.Success)
		fmt.Printf("Final Data: %+v\n", result.Data)
	}
}
