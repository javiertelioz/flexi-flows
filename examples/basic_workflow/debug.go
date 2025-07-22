package main

import (
	"context"
	"fmt"
	"log"

	"github.com/javiertelioz/flexi-flows/examples/basic_workflow/tasks"
)

func main() {
	fmt.Println("🧪 Testing individual task execution...")

	ctx := context.Background()
	inputData := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30.0,
		},
	}

	fmt.Printf("Input Data: %+v\n", inputData)

	// Test ValidateData task directly
	result, err := tasks.ValidateData(ctx, inputData)
	if err != nil {
		log.Fatalf("ValidateData failed: %v", err)
	}

	fmt.Printf("✅ ValidateData result: %+v\n", result)

	// Test TransformData task
	result, err = tasks.TransformData(ctx, result)
	if err != nil {
		log.Fatalf("TransformData failed: %v", err)
	}

	fmt.Printf("✅ TransformData result: %+v\n", result)

	fmt.Println("🎉 Individual tasks work correctly!")
}
