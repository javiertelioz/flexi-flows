package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/javiertelioz/flexi-flows/examples/conditional_branching/programmatic"
	"github.com/javiertelioz/flexi-flows/examples/conditional_branching/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

func main() {
	// Parse command line flags
	mode := flag.String("mode", "config", "Execution mode: config, programmatic, or autodiscovery")
	customer := flag.String("customer", "regular", "Customer type: vip, regular, new, invalid")
	orderValue := flag.Float64("order", 250.0, "Order value")
	flag.Parse()

	// Sample input data based on customer type
	inputData := map[string]interface{}{
		"order": map[string]interface{}{
			"id":            "ORD_" + *customer + "_001",
			"customer_id":   "CUST_" + *customer + "_123",
			"customer_type": *customer,
			"order_value":   *orderValue,
			"items": []map[string]interface{}{
				{"name": "Product A", "quantity": 2, "price": *orderValue / 2},
			},
			"shipping_address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "New York",
				"zip":    "10001",
			},
		},
	}

	fmt.Printf("🚀 Starting Conditional Branching Example\n")
	fmt.Printf("📋 Mode: %s\n", *mode)
	fmt.Printf("👤 Customer Type: %s\n", *customer)
	fmt.Printf("💰 Order Value: $%.2f\n\n", *orderValue)

	var result *workflow.ExecutionResult
	var err error

	switch *mode {
	case "programmatic":
		result, err = runProgrammaticMode(inputData)
	case "config":
		result, err = runConfigMode(inputData)
	case "autodiscovery":
		result, err = runAutoDiscoveryMode(inputData)
	default:
		log.Fatalf("❌ Unknown mode: %s. Use: config, programmatic, or autodiscovery", *mode)
	}

	if err != nil {
		log.Fatalf("❌ Conditional branching execution failed: %v", err)
	}

	// Display results
	fmt.Printf("✅ Workflow completed successfully!\n")
	fmt.Printf("\n🎉 Conditional Branching completed!\n")
	fmt.Printf("📈 Execution Details:\n")
	fmt.Printf("  - Node ID: %s\n", result.NodeID)
	fmt.Printf("  - Success: %t\n", result.Success)
	fmt.Printf("  - Timestamp: %d\n", result.Timestamp)

	// Pretty print result
	if result.Data != nil {
		jsonResult, _ := json.MarshalIndent(result.Data, "", "  ")
		fmt.Printf("\n📋 Final Result:\n%s\n", jsonResult)
	}
}

func runProgrammaticMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔧 Running in PROGRAMMATIC mode...")
	return programmatic.ExecuteConditionalBranchingWorkflow(inputData)
}

func runConfigMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("⚙️ Running in CONFIG mode...")

	wm := workflow.NewWorkflowManager()

	// Register tasks
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

	// Load workflow from JSON config
	err := wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Execute workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "classifyOrder", inputData)
}

func runAutoDiscoveryMode(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔍 Running in AUTO-DISCOVERY mode...")

	wm := workflow.NewWorkflowManager()

	// Create auto-discoverer
	autodiscoverer := workflow.NewAutoDiscoverer()

	// Scan the tasks package for functions with annotations
	err := autodiscoverer.ScanPackage("./tasks")
	if err != nil {
		return nil, fmt.Errorf("failed to scan tasks package: %w", err)
	}

	// Register discovered functions with the workflow manager
	err = autodiscoverer.RegisterInManager(wm)
	if err != nil {
		return nil, fmt.Errorf("failed to register discovered functions: %w", err)
	}

	// For now, use the same config as the config mode since auto-generation is complex
	err = wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config for auto-discovery: %w", err)
	}

	// Execute workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "classifyOrder", inputData)
}
