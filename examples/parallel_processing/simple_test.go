package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

// Función simple para simular obtención de datos del clima
func fetchWeatherData(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("🌦️ Fetching weather data...")

	citiesStr, ok := data["cities"].(string)
	if !ok {
		citiesStr = "New York,London,Tokyo"
	}

	cities := strings.Split(citiesStr, ",")
	weatherData := make([]interface{}, 0)

	for _, city := range cities {
		city = strings.TrimSpace(city)

		// Simular delay de API
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

		weather := map[string]interface{}{
			"city":        city,
			"temperature": rand.Float64()*35 + 5, // 5-40°C
			"humidity":    rand.Float64() * 100,  // 0-100%
			"condition":   []string{"sunny", "cloudy", "rainy", "windy"}[rand.Intn(4)],
		}

		weatherData = append(weatherData, weather)
		fmt.Printf("📍 Fetched weather for %s: %.1f°C\n", city, weather["temperature"])
	}

	result := map[string]interface{}{
		"weather_data": weatherData,
		"total_cities": len(cities),
	}

	fmt.Printf("✅ Successfully fetched weather data for %d cities\n", len(cities))
	return result, nil
}

// Función simple para agregar datos del clima
func aggregateWeatherData(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("📊 Aggregating weather data...")

	weatherData, ok := data["weather_data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no weather data found")
	}

	var totalTemp, totalHumidity float64
	conditionCount := make(map[string]int)

	for _, item := range weatherData {
		weather := item.(map[string]interface{})
		totalTemp += weather["temperature"].(float64)
		totalHumidity += weather["humidity"].(float64)
		conditionCount[weather["condition"].(string)]++
	}

	avgTemp := totalTemp / float64(len(weatherData))
	avgHumidity := totalHumidity / float64(len(weatherData))

	// Encontrar condición más común
	var mostCommonCondition string
	maxCount := 0
	for condition, count := range conditionCount {
		if count > maxCount {
			maxCount = count
			mostCommonCondition = condition
		}
	}

	aggregated := map[string]interface{}{
		"total_cities":          len(weatherData),
		"average_temperature":   avgTemp,
		"average_humidity":      avgHumidity,
		"most_common_condition": mostCommonCondition,
	}

	fmt.Printf("📈 Aggregated %d cities: Avg temp %.1f°C, Most common: %s\n",
		len(weatherData), avgTemp, mostCommonCondition)

	return map[string]interface{}{
		"aggregated_data": aggregated,
		"raw_data":        data,
	}, nil
}

func main() {
	fmt.Println("🚀 Starting Simple Weather Processing Example")

	// Crear el workflow manager
	wm := workflow.NewWorkflowManager()

	// Registrar las tareas
	wm.RegisterTask("fetchWeatherData", fetchWeatherData)
	wm.RegisterTask("aggregateWeatherData", aggregateWeatherData)

	// Crear configuración de workflow
	cfg := &config.WorkflowConfig{
		Name:        "simple_weather_workflow",
		Description: "Simple weather processing workflow",
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

	// Construir el workflow desde la configuración
	err := wm.BuildFromConfig(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to build workflow: %v", err)
	}

	// Datos de entrada
	inputData := map[string]interface{}{
		"cities": "Madrid,Barcelona,Valencia,Sevilla",
	}

	// Ejecutar el workflow
	fmt.Println("🔧 Executing workflow...")
	ctx := context.Background()
	result, err := wm.ExecuteWithContext(ctx, "fetch_weather", inputData)

	if err != nil {
		log.Fatalf("❌ Workflow execution failed: %v", err)
	}

	// Mostrar resultados
	fmt.Println("\n🎉 Workflow completed successfully!")
	fmt.Printf("📋 Node ID: %s\n", result.NodeID)
	fmt.Printf("⏱️ Duration: %d ms\n", result.Duration)
	fmt.Printf("✅ Success: %t\n", result.Success)

	if result.Data != nil {
		if dataMap, ok := result.Data.(map[string]interface{}); ok {
			if aggregated, ok := dataMap["aggregated_data"].(map[string]interface{}); ok {
				fmt.Println("\n🌤️ Weather Summary:")
				fmt.Printf("  • Cities processed: %v\n", aggregated["total_cities"])
				fmt.Printf("  • Average temperature: %.1f°C\n", aggregated["average_temperature"])
				fmt.Printf("  • Average humidity: %.1f%%\n", aggregated["average_humidity"])
				fmt.Printf("  • Most common condition: %s\n", aggregated["most_common_condition"])
			}
		}
	}
}
