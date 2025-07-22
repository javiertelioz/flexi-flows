package tasks

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// @workflow:task
// @name: fetchWeatherData
// @description: Fetches weather data for multiple cities in parallel
func FetchWeatherData(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🌦️ Starting weather data fetch...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	citiesStr, _ := dataMap["cities"].(string)
	if citiesStr == "" {
		citiesStr = "New York,London,Tokyo"
	}

	cities := strings.Split(citiesStr, ",")
	weatherData := make([]interface{}, len(cities))

	// Simulate parallel API calls
	for i, city := range cities {
		city = strings.TrimSpace(city)

		// Simulate API delay
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		weather := map[string]interface{}{
			"city":        city,
			"temperature": rand.Float64()*35 + 5, // 5-40°C
			"humidity":    rand.Float64() * 100,  // 0-100%
			"condition":   []string{"sunny", "cloudy", "rainy", "windy"}[rand.Intn(4)],
			"timestamp":   time.Now().Unix(),
		}
		weatherData[i] = weather

		fmt.Printf("📍 Fetched weather for %s: %.1f°C, %s\n",
			city, weather["temperature"], weather["condition"])
	}

	result := map[string]interface{}{
		"weather_data": weatherData,
		"total_cities": len(cities),
		"fetch_time":   time.Now().Unix(),
	}

	fmt.Printf("✅ Successfully fetched weather data for %d cities\n", len(cities))
	return result, nil
}

// @workflow:task
// @name: aggregateWeatherData
// @description: Aggregates weather data from multiple cities
func AggregateWeatherData(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("📊 Starting weather data aggregation...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	weatherData, ok := dataMap["weather_data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no weather data found to aggregate")
	}

	if len(weatherData) == 0 {
		return map[string]interface{}{
			"aggregated_data": map[string]interface{}{
				"total_cities":        0,
				"average_temperature": 0.0,
				"average_humidity":    0.0,
				"condition_summary":   "no data",
			},
		}, nil
	}

	var totalTemp, totalHumidity float64
	conditionCount := make(map[string]int)

	// Process each city's weather data
	for _, item := range weatherData {
		weather, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if temp, ok := weather["temperature"].(float64); ok {
			totalTemp += temp
		}
		if humidity, ok := weather["humidity"].(float64); ok {
			totalHumidity += humidity
		}
		if condition, ok := weather["condition"].(string); ok {
			conditionCount[condition]++
		}
	}

	// Calculate averages
	cityCount := float64(len(weatherData))
	avgTemp := totalTemp / cityCount
	avgHumidity := totalHumidity / cityCount

	// Find most common condition
	var mostCommonCondition string
	maxCount := 0
	for condition, count := range conditionCount {
		if count > maxCount {
			maxCount = count
			mostCommonCondition = condition
		}
	}

	aggregatedData := map[string]interface{}{
		"total_cities":        len(weatherData),
		"average_temperature": avgTemp,
		"average_humidity":    avgHumidity,
		"condition_summary":   mostCommonCondition,
		"condition_breakdown": conditionCount,
		"processed_at":        time.Now().Unix(),
	}

	fmt.Printf("📊 Aggregated data for %d cities:\n", len(weatherData))
	fmt.Printf("   - Average Temperature: %.1f°C\n", avgTemp)
	fmt.Printf("   - Average Humidity: %.1f%%\n", avgHumidity)
	fmt.Printf("   - Most Common Condition: %s\n", mostCommonCondition)

	return map[string]interface{}{
		"aggregated_data": aggregatedData,
		"raw_data":        dataMap,
	}, nil
}
