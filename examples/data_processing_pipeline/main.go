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
		log.Fatalf("❌ Data processing pipeline execution failed: %v", err)
	}

	// Display results
	fmt.Printf("✅ Workflow completed successfully!\n")
	fmt.Printf("\n🎉 Data Processing Pipeline completed!\n")
	fmt.Printf("📈 Execution Details:\n")
	fmt.Printf("  - Node ID: %s\n", result.NodeID)
	fmt.Printf("  - Success: %t\n", result.Success)
	fmt.Printf("  - Timestamp: %d\n", result.Timestamp)

	// Pretty print result
	if result.Data != nil {
		jsonResult, _ := json.MarshalIndent(result.Data, "", "  ")
		fmt.Printf("\n📋 Final Result:\n%s\n", jsonResult)
	}
}

func runProgrammaticMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔧 Running in PROGRAMMATIC mode...")
	return programmatic.ExecuteDataProcessingPipeline(inputData)
}

func runConfigMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("⚙️ Running in CONFIG mode...")

	wm := workflow.NewWorkflowManager()

	// Register tasks
	wm.RegisterTask("extractData", tasks.ExtractData)
	wm.RegisterTask("validateData", tasks.ValidateData)
	wm.RegisterTask("transformData", tasks.TransformData)
	wm.RegisterTask("enrichData", tasks.EnrichData)
	wm.RegisterTask("filterData", tasks.FilterData)
	wm.RegisterTask("aggregateData", tasks.AggregateData)
	wm.RegisterTask("saveData", tasks.SaveData)

	// Load workflow from config
	err := wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Execute workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "extractData", inputData)
}

func runAutoDiscoveryMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔍 Running in AUTO-DISCOVERY mode...")

	wm := workflow.NewWorkflowManager()

	// First, register tasks manually to ensure they're available
	fmt.Println("📝 Registering tasks manually as fallback...")
	wm.RegisterTask("extractData", tasks.ExtractData)
	wm.RegisterTask("validateData", tasks.ValidateData)
	wm.RegisterTask("transformData", tasks.TransformData)
	wm.RegisterTask("enrichData", tasks.EnrichData)
	wm.RegisterTask("filterData", tasks.FilterData)
	wm.RegisterTask("aggregateData", tasks.AggregateData)
	wm.RegisterTask("saveData", tasks.SaveData)
	fmt.Println("✅ Manual task registration completed")

	// Then try auto-discovery as an additional feature
	fmt.Println("🔍 Attempting auto-discovery...")
	autodiscoverer := workflow.NewAutoDiscoverer()

	// Scan the tasks package for functions with annotations
	err := autodiscoverer.ScanPackage("./tasks")
	if err != nil {
		fmt.Printf("⚠️  Auto-discovery scanning failed: %v\n", err)
	} else {
		fmt.Println("✅ Auto-discovery scan completed")
	}

	// Register discovered functions with the workflow manager
	err = autodiscoverer.RegisterInManager(wm)
	if err != nil {
		fmt.Printf("⚠️  Auto-discovery registration failed: %v\n", err)
	} else {
		fmt.Println("✅ Auto-discovery registration completed")
	}

	// Load workflow from config
	fmt.Println("📋 Loading configuration...")
	err = wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	fmt.Println("✅ Configuration loaded successfully")

	// Execute workflow
	fmt.Println("▶️  Starting workflow execution...")
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "extractData", inputData)
}
