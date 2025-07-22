package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("🚀 Starting Data Processing Pipeline Debug Test")
	log.Println("Debug: Starting...")

	// Test 1: Basic functionality
	fmt.Println("\n1️⃣ Testing basic functionality...")

	inputData := map[string]interface{}{
		"source":     "mock",
		"batch_size": 3,
		"processing_rules": map[string]interface{}{
			"validate_email": true,
		},
	}

	fmt.Printf("📋 Input data: %+v\n", inputData)
	log.Printf("Debug: Input data created successfully")

	fmt.Println("✅ Basic test completed!")
	log.Println("Debug: Test completed!")
}
