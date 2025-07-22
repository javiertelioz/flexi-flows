package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/javiertelioz/flexi-flows/examples/parallel_processing/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

func main() {
	// Parse command line flags
	mode := flag.String("mode", "config", "Execution mode: config or programmatic")
	cities := flag.String("cities", "New York,London,Tokyo", "Cities to get weather for (comma-separated)")
	flag.Parse()

	// Input data
	inputData := map[string]interface{}{
		"cities": *cities,
	}

	fmt.Printf("🚀 Starting Weather Processing Example\n")
	fmt.Printf("📋 Mode: %s\n", *mode)
	fmt.Printf("🌍 Cities: %s\n\n", *cities)

	var result *workflow.ExecutionResult
	var err error

	switch *mode {
	case "config":
		result, err = runConfigMode(inputData)
	case "programmatic":
		result, err = runProgrammaticMode(inputData)
	default:
		log.Fatalf("❌ Unknown mode: %s", *mode)
	}

	if err != nil {
		log.Fatalf("❌ Workflow execution failed: %v", err)
	}

	// Display results
	fmt.Printf("✅ Config workflow completed successfully!\n")
	fmt.Printf("\n🎉 Workflow completed successfully!\n")
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

func runConfigMode(inputData map[string]interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("⚙️ Running in CONFIG mode...")

	wm := workflow.NewWorkflowManager()

	// Register tasks
	wm.RegisterTask("fetchWeatherData", tasks.FetchWeatherData)
	wm.RegisterTask("aggregateWeatherData", tasks.AggregateWeatherData)

	// Load configuration
	err := wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Execute workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "fetch_weather", inputData)
}

func runProgrammaticMode(inputData map[string]interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔧 Running in PROGRAMMATIC mode...")

	wm := workflow.NewWorkflowManager()

	// Register tasks
	wm.RegisterTask("fetchWeatherData", tasks.FetchWeatherData)
	wm.RegisterTask("aggregateWeatherData", tasks.AggregateWeatherData)

	// Create workflow configuration programmatically
	cfg := &config.WorkflowConfig{
		Name:        "weather_processing_workflow",
		Description: "Weather data processing workflow",
		Version:     "1.0.0",
		StartNode:   "fetch_weather",
		Nodes: []config.NodeConfig{
			{
				ID:       "fetch_weather",
				Type:     "task",
				Function: "fetchWeatherData",
				Next:     []string{"aggregate_weather"},
			},
			{
				ID:       "aggregate_weather",
				Type:     "task",
				Function: "aggregateWeatherData",
			},
		},
	}

	// Build workflow from config
	err := wm.BuildFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build programmatic workflow: %w", err)
	}

	// Execute workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "fetch_weather", inputData)
}
