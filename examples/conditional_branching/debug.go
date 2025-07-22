package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/javiertelioz/flexi-flows/examples/conditional_branching/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

func main() {
	fmt.Println("🚀 Starting Conditional Branching Debug Test")

	// Simple test data
	inputData := map[string]interface{}{
		"order": map[string]interface{}{
			"id":            "ORD_vip_001",
			"customer_id":   "CUST_vip_123",
			"customer_type": "vip",
			"order_value":   500.0,
		},
	}

	fmt.Println("📋 Input data:", inputData)

	// Test 1: Direct task execution
	fmt.Println("\n1️⃣ Testing direct task execution...")
	ctx := context.Background()

	result, err := tasks.ClassifyOrder(ctx, inputData)
	if err != nil {
		log.Printf("❌ Classification failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ Classification successful:", result)

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
	wm.RegisterTask("classifyOrder", tasks.ClassifyOrder)
	wm.RegisterTask("processVIPOrder", tasks.ProcessVIPOrder)
	wm.RegisterTask("processRegularOrder", tasks.ProcessRegularOrder)
	wm.RegisterTask("processNewCustomer", tasks.ProcessNewCustomer)
	wm.RegisterTask("handleInvalidOrder", tasks.HandleInvalidOrder)
	wm.RegisterTask("finalizeOrder", tasks.FinalizeOrder)

	// Register conditional functions
	wm.RegisterTask("isVIPCustomer", tasks.IsVIPCustomer)
	wm.RegisterTask("isRegularCustomer", tasks.IsRegularCustomer)
	wm.RegisterTask("isNewCustomer", tasks.IsNewCustomer)

	fmt.Println("✅ Tasks and conditionals registered successfully")

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
	executionResult, err := wm.ExecuteWithContext(ctx, "classifyOrder", inputData)
	if err != nil {
		log.Printf("❌ Workflow execution failed: %v", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Workflow execution completed successfully!\n")
	fmt.Printf("📈 Final result: %+v\n", executionResult)
	fmt.Println("🎉 All debug tests passed!")
}
