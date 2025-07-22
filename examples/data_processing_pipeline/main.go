package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/javiertelioz/flexi-flows/examples/data_processing_pipeline/programmatic"
	"github.com/javiertelioz/flexi-flows/examples/data_processing_pipeline/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

func main() {
	// Parse command line flags
	mode := flag.String("mode", "config", "Execution mode: config, programmatic, or autodiscovery")
	source := flag.String("source", "mock", "Data source: api or mock")
	flag.Parse()

	// Input data with configuration
	inputData := map[string]interface{}{
		"source":     *source,
		"batch_size": 5,
		"processing_rules": map[string]interface{}{
			"validate_email":   true,
			"normalize_names":  true,
			"extract_domains":  true,
			"calculate_scores": true,
		},
	}

	fmt.Printf("🚀 Starting Data Processing Pipeline Example\n")
	fmt.Printf("📋 Mode: %s\n", *mode)
	fmt.Printf("🌐 Source: %s\n", *source)
	fmt.Printf("📊 Processing Rules: %+v\n\n", inputData["processing_rules"])

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
		log.Fatalf("❌ Data pipeline execution failed: %v", err)
	}

	// Pretty print results
	fmt.Printf("\n🎉 Data Processing Pipeline completed successfully!\n")
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

	// Extract and display key metrics
	if result.Data != nil {
		if dataMap, ok := result.Data.(map[string]interface{}); ok {
			fmt.Printf("\n📊 Processing Summary:\n")

			if summary, ok := dataMap["processing_summary"].(map[string]interface{}); ok {
				fmt.Printf("  - Total Users: %v\n", summary["total_users"])
				fmt.Printf("  - Valid Users: %v\n", summary["valid_users"])
				fmt.Printf("  - Invalid Users: %v\n", summary["invalid_users"])
				fmt.Printf("  - Saved Users: %v\n", summary["saved_users"])
				fmt.Printf("  - Domains Found: %v\n", summary["domains_found"])
				fmt.Printf("  - Batch ID: %v\n", summary["batch_id"])
			}

			if validationErrors, ok := dataMap["validation_errors"].([]interface{}); ok && len(validationErrors) > 0 {
				fmt.Printf("\n⚠️  Validation Issues:\n")
				for _, err := range validationErrors {
					fmt.Printf("  - %v\n", err)
				}
			}
		}

		// Pretty print full result for debugging (commented out by default)
		// jsonResult, _ := json.MarshalIndent(result.Data, "", "  ")
		// fmt.Printf("\n📋 Full Result:\n%s\n", jsonResult)
	}
}

func runProgrammaticMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔧 Running in PROGRAMMATIC mode...")
	return programmatic.ExecuteDataPipelineWorkflow(inputData)
}

func runConfigMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("⚙️ Running in CONFIG mode...")

	wm := workflow.NewWorkflowManager()

	// Register tasks
	wm.RegisterTask("fetchUsers", tasks.FetchUsers)
	wm.RegisterTask("validateUsers", tasks.ValidateUsers)
	wm.RegisterTask("transformUsers", tasks.TransformUsers)
	wm.RegisterTask("saveUsers", tasks.SaveUsers)
	wm.RegisterTask("generateReport", tasks.GenerateReport)

	// Load workflow from JSON config
	err := wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	result, err := wm.ExecuteWithContext(ctx, "fetch", inputData)

	if err != nil {
		return nil, fmt.Errorf("config workflow execution failed: %w", err)
	}

	fmt.Println("✅ Config workflow completed successfully!")
	return result, nil
}

func runAutoDiscoveryMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔍 Running in AUTO-DISCOVERY mode...")

	wm := workflow.NewWorkflowManager()

	// Register tasks manually (same strategy as basic_workflow)
	wm.RegisterTask("fetchUsers", tasks.FetchUsers)
	wm.RegisterTask("validateUsers", tasks.ValidateUsers)
	wm.RegisterTask("transformUsers", tasks.TransformUsers)
	wm.RegisterTask("saveUsers", tasks.SaveUsers)
	wm.RegisterTask("generateReport", tasks.GenerateReport)

	fmt.Printf("🔍 Auto-discovered (manual) tasks: fetchUsers, validateUsers, transformUsers, saveUsers, generateReport\n")

	// Create workflow configuration using the discovered tasks
	cfg := &config.WorkflowConfig{
		Name:        "data_pipeline_autodiscovery_workflow",
		Description: "Data processing pipeline with auto-discovery",
		Version:     "1.0.0",
		StartNode:   "fetch",
		Nodes: []config.NodeConfig{
			{
				ID:       "fetch",
				Type:     "task",
				Function: "fetchUsers",
			},
			{
				ID:       "validate",
				Type:     "task",
				Function: "validateUsers",
			},
			{
				ID:       "transform",
				Type:     "task",
				Function: "transformUsers",
			},
			{
				ID:       "save",
				Type:     "task",
				Function: "saveUsers",
			},
			{
				ID:       "report",
				Type:     "task",
				Function: "generateReport",
			},
		},
		Edges: []config.EdgeConfig{
			{From: "fetch", To: "validate"},
			{From: "validate", To: "transform"},
			{From: "transform", To: "save"},
			{From: "save", To: "report"},
		},
	}

	// Build workflow from config
	err := wm.BuildFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build auto-discovery workflow: %w", err)
	}

	ctx := context.Background()
	result, err := wm.ExecuteWithContext(ctx, "fetch", inputData)

	if err != nil {
		return nil, fmt.Errorf("auto-discovery workflow execution failed: %w", err)
	}

	fmt.Println("✅ Auto-discovery workflow completed successfully!")
	return result, nil
}
