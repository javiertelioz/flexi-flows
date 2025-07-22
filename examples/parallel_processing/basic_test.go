package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("🚀 Testing basic execution...")

	// Simple function test without workflows
	result := simpleWeatherFunction("Madrid,Barcelona")

	fmt.Printf("✅ Result: %+v\n", result)
}

func simpleWeatherFunction(cities string) map[string]interface{} {
	fmt.Printf("📍 Processing cities: %s\n", cities)

	return map[string]interface{}{
		"cities": cities,
		"status": "processed",
	}
}
