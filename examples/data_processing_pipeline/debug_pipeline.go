package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/javiertelioz/flexi-flows/examples/data_processing_pipeline/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

func main() {
	fmt.Println("🚀 Starting Data Processing Pipeline Debug Test")

	// Simple test data
	inputData := map[string]interface{}{
		"source":     "mock",
		"batch_size": 3,
	}

	fmt.Println("📋 Input data:", inputData)

	// Test 1: Direct task execution
	fmt.Println("\n1️⃣ Testing direct task execution...")
	ctx := context.Background()

	result, err := tasks.ExtractData(ctx, inputData)
	if err != nil {
		log.Printf("❌ ExtractData failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ ExtractData successful")

	// Test 2: Workflow manager creation
	fmt.Println("\n2️⃣ Testing workflow manager creation...")
	wm := workflow.NewWorkflowManager()
	if wm == nil {
		log.Println("❌ Failed to create workflow manager")
		os.Exit(1)
	}
	fmt.Println("✅ Workflow manager created successfully")

	// Test 3: Task registration
	fmt.Println("\n3️⃣ Testing task registration...")
	wm.RegisterTask("extractData", tasks.ExtractData)
	wm.RegisterTask("validateData", tasks.ValidateData)
	wm.RegisterTask("transformData", tasks.TransformData)
	wm.RegisterTask("enrichData", tasks.EnrichData)
	wm.RegisterTask("filterData", tasks.FilterData)
	wm.RegisterTask("aggregateData", tasks.AggregateData)
	wm.RegisterTask("saveData", tasks.SaveData)

	fmt.Println("✅ Tasks registered successfully")

	// Test 4: Config loading
	fmt.Println("\n4️⃣ Testing config loading...")
	err = wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		log.Printf("❌ Config loading failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ Config loaded successfully")

	// Test 5: Workflow execution
	fmt.Println("\n5️⃣ Testing workflow execution...")
	executionResult, err := wm.ExecuteWithContext(ctx, "extractData", inputData)
	if err != nil {
		log.Printf("❌ Workflow execution failed: %v", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Workflow execution completed successfully!\n")
	fmt.Printf("📈 Final result: %+v\n", executionResult)
	fmt.Println("🎉 All debug tests passed!")
}
