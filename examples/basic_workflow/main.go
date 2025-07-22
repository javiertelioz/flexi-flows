package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/javiertelioz/flexi-flows/examples/basic_workflow/programmatic"
	"github.com/javiertelioz/flexi-flows/examples/basic_workflow/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

func main() {
	// Parse command line flags
	mode := flag.String("mode", "config", "Execution mode: config, programmatic, or autodiscovery")
	flag.Parse()

	// Sample input data
	inputData := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30.0,
		},
	}

	fmt.Printf("🚀 Starting Basic Workflow Example\n")
	fmt.Printf("📋 Mode: %s\n", *mode)
	fmt.Printf("📊 Input Data: %+v\n\n", inputData)

	var result *workflow.ExecutionResult

	var err error

	switch *mode {
	case "programmatic":
		result, err = runProgrammaticMode(inputData)
	case "config":
		result, err = runConfigMode(inputData)
	case "autodiscovery":
		result, err = runAutoDiscoveryMode(inputData)
	default:
		log.Fatalf("❌ Unknown mode: %s. Use: config, programmatic, or autodiscovery", *mode)
	}

	if err != nil {
		log.Fatalf("❌ Workflow execution failed: %v", err)
	}

	// Pretty print results
	fmt.Printf("\n🎉 Workflow completed successfully!\n")
	fmt.Printf("📈 Execution Details:\n")
	if result.NodeID != "" {
		fmt.Printf("  - Node ID: %s\n", result.NodeID)
	}
	if result.Duration > 0 {
		fmt.Printf("  - Duration: %d ms\n", result.Duration)
	}
	fmt.Printf("  - Success: %t\n", result.Success)
	if result.Timestamp > 0 {
		fmt.Printf("  - Timestamp: %d\n", result.Timestamp)
	}

	// Pretty print final data
	if result.Data != nil {
		jsonResult, _ := json.MarshalIndent(result.Data, "", "  ")
		fmt.Printf("\n📋 Final Result:\n%s\n", jsonResult)
	}
}

func runProgrammaticMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔧 Running in PROGRAMMATIC mode...")
	return programmatic.ExecuteProgrammaticWorkflow(inputData)
}

func runConfigMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("⚙️ Running in CONFIG mode...")

	wm := workflow.NewWorkflowManager()

	// Register tasks
	wm.RegisterTask("validateData", tasks.ValidateData)
	wm.RegisterTask("transformData", tasks.TransformData)
	wm.RegisterTask("saveData", tasks.SaveData)
	wm.RegisterTask("notifyUser", tasks.NotifyUser)

	// Load workflow from JSON config
	err := wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	result, err := wm.ExecuteWithContext(ctx, "validate", inputData)

	if err != nil {
		return nil, fmt.Errorf("config workflow execution failed: %w", err)
	}

	fmt.Println("✅ Config workflow completed successfully!")
	return result, nil
}

func runAutoDiscoveryMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔍 Running in AUTO-DISCOVERY mode...")

	wm := workflow.NewWorkflowManager()

	// Para este ejemplo, registramos las tareas manualmente
	// En una implementación futura completa, el auto-discovery funcionaría automáticamente
	wm.RegisterTask("validateData", tasks.ValidateData)
	wm.RegisterTask("transformData", tasks.TransformData)
	wm.RegisterTask("saveData", tasks.SaveData)
	wm.RegisterTask("notifyUser", tasks.NotifyUser)

	fmt.Printf("🔍 Auto-discovered (manual) tasks: validateData, transformData, saveData, notifyUser\n")

	// Create workflow configuration using the discovered tasks
	cfg := &config.WorkflowConfig{
		Name:        "basic_autodiscovery_workflow",
		Description: "Basic workflow with auto-discovery",
		Version:     "1.0.0",
		StartNode:   "validate",
		Nodes: []config.NodeConfig{
			{
				ID:       "validate",
				Type:     "task",
				Function: "validateData",
			},
			{
				ID:       "transform",
				Type:     "task",
				Function: "transformData",
			},
			{
				ID:       "save",
				Type:     "task",
				Function: "saveData",
			},
			{
				ID:       "notify",
				Type:     "task",
				Function: "notifyUser",
			},
		},
		Edges: []config.EdgeConfig{
			{From: "validate", To: "transform"},
			{From: "transform", To: "save"},
			{From: "save", To: "notify"},
		},
	}

	// Build workflow from config
	err := wm.BuildFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build auto-discovery workflow: %w", err)
	}

	ctx := context.Background()
	result, err := wm.ExecuteWithContext(ctx, "validate", inputData)

	if err != nil {
		return nil, fmt.Errorf("auto-discovery workflow execution failed: %w", err)
	}

	fmt.Println("✅ Auto-discovery workflow completed successfully!")
	return result, nil
}
