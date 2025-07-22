package main

import (
	"context"
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

	// Input data with cities configuration
	inputData := map[string]interface{}{
		"cities": *cities,
		"config": map[string]interface{}{
			"timeout":     "30s",
			"max_retries": 3,
		},
	}

	fmt.Printf("🚀 Starting Parallel Processing Example\n")
	fmt.Printf("📋 Mode: %s\n", *mode)
	fmt.Printf("🌍 Cities: %s\n", *cities)
	fmt.Printf("📊 Config: %+v\n\n", inputData["config"])

	var result *workflow.ExecutionResult
	var err error

	switch *mode {
	case "programmatic":
		result, err = runProgrammaticMode(inputData)
	case "config":
		result, err = runConfigMode(inputData)
	default:
		log.Fatalf("❌ Unknown mode: %s. Use: config or programmatic", *mode)
	}

	if err != nil {
		log.Fatalf("❌ Parallel processing execution failed: %v", err)
	}

	// Pretty print results
	fmt.Printf("\n🎉 Parallel Processing completed successfully!\n")
	fmt.Printf("📈 Execution Details:\n")
	if result != nil {
		if result.NodeID != "" {
			fmt.Printf("  - Node ID: %s\n", result.NodeID)
		}
		if result.Duration > 0 {
			fmt.Printf("  - Duration: %d ms\n", result.Duration)
		}
		fmt.Printf("  - Success: %t\n", result.Success)

		// Extract and display weather summary
		if result.Data != nil {
			if dataMap, ok := result.Data.(map[string]interface{}); ok {
				fmt.Printf("\n🌤️ Weather Summary:\n")
				if summary, ok := dataMap["aggregated_data"].(map[string]interface{}); ok {
					if avgTemp, ok := summary["average_temperature"].(float64); ok {
						fmt.Printf("  - Average Temperature: %.1f°C\n", avgTemp)
					}
					if avgHumidity, ok := summary["average_humidity"].(float64); ok {
						fmt.Printf("  - Average Humidity: %.1f%%\n", avgHumidity)
					}
					if totalCities, ok := summary["total_cities"].(int); ok {
						fmt.Printf("  - Total Cities: %d\n", totalCities)
					}
					if condition, ok := summary["condition_summary"].(string); ok {
						fmt.Printf("  - Most Common Condition: %s\n", condition)
					}
				}
			}
		}
	}
}

func runProgrammaticMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔧 Running in PROGRAMMATIC mode...")

	wm := workflow.NewWorkflowManager()

	// Register tasks
	wm.RegisterTask("fetchWeatherData", tasks.FetchWeatherData)
	wm.RegisterTask("aggregateWeatherData", tasks.AggregateWeatherData)

	// Create workflow configuration programmatically
	cfg := &config.WorkflowConfig{
		Name:        "parallel_weather_programmatic",
		Description: "Parallel weather processing created programmatically",
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
	result, err := wm.ExecuteWithContext(ctx, "fetch_weather", inputData)

	if err != nil {
		return nil, fmt.Errorf("programmatic workflow execution failed: %w", err)
	}

	fmt.Println("✅ Programmatic workflow completed successfully!")
	return result, nil
}

func runConfigMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("⚙️ Running in CONFIG mode...")

	wm := workflow.NewWorkflowManager()

	// Register tasks
	wm.RegisterTask("fetchWeatherData", tasks.FetchWeatherData)
	wm.RegisterTask("aggregateWeatherData", tasks.AggregateWeatherData)

	// Load workflow from JSON config
	err := wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	result, err := wm.ExecuteWithContext(ctx, "fetch_weather", inputData)

	if err != nil {
		return nil, fmt.Errorf("config workflow execution failed: %w", err)
	}

	fmt.Println("✅ Config workflow completed successfully!")
	return result, nil
}
