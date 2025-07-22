package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

import (
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

// Simple weather fetch function
func fetchWeather(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🌦️ Fetching weather data...")

	dataMap := data.(map[string]interface{})
	cities := strings.Split(dataMap["cities"].(string), ",")

	results := make([]interface{}, 0)
	for _, city := range cities {
		city = strings.TrimSpace(city)
		weather := map[string]interface{}{
			"city":        city,
			"temperature": rand.Float64()*30 + 10,
			"condition":   "sunny",
		}
		results = append(results, weather)
		fmt.Printf("📍 %s: %.1f°C\n", city, weather["temperature"])
	}

	return map[string]interface{}{
		"weather_data": results,
		"total":        len(results),
	}, nil
}

func main() {
	fmt.Println("🚀 Testing Simple Weather Workflow")

	// Create workflow manager
	wm := workflow.NewWorkflowManager()

	// Register the task
	wm.RegisterTask("fetchWeather", fetchWeather)

	// Create a simple task node
	taskNode := &workflow.TaskNode{
		BaseNode: workflow.BaseNode{
			ID:   "weather_task",
			Type: "task",
		},
		Function: "fetchWeather",
	}

	// Add node to workflow
	wm.AddNode(taskNode)

	// Input data
	inputData := map[string]interface{}{
		"cities": "Madrid,Barcelona,Valencia",
	}

	// Execute
	ctx := context.Background()
	result, err := wm.ExecuteWithContext(ctx, "weather_task", inputData)

	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("\n✅ Success: %t\n", result.Success)
	fmt.Printf("📋 Node: %s\n", result.NodeID)

	if result.Data != nil {
		if dataMap, ok := result.Data.(map[string]interface{}); ok {
			if total, ok := dataMap["total"].(int); ok {
				fmt.Printf("📊 Processed %d cities\n", total)
			}
		}
	}
}
