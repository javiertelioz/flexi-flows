package main

import (
	"fmt"
)

func main() {
	fmt.Println("🚀 Basic Test - Starting...")
	fmt.Println("✅ If you see this, basic Go is working")

	// Test basic data creation
	testData := map[string]interface{}{
		"source": "mock",
		"test":   true,
	}

	fmt.Printf("📋 Test data: %+v\n", testData)
	fmt.Println("🎉 Basic test completed successfully!")
}
