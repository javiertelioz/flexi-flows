package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/javiertelioz/flexi-flows/examples/microservices_orchestration/programmatic"
	"github.com/javiertelioz/flexi-flows/examples/microservices_orchestration/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

func main() {
	fmt.Println("🚀 Starting Microservices Orchestration Example")

	// Command line flags
	mode := flag.String("mode", "programmatic", "Execution mode: programmatic, config, or autodiscovery")
	orderID := flag.String("order", "ORD-2024-001", "Order ID to process")
	userID := flag.String("user", "user123", "User ID")
	flag.Parse()

	fmt.Printf("📋 Mode: %s\n", *mode)
	fmt.Printf("🛒 Order ID: %s\n", *orderID)
	fmt.Printf("👤 User ID: %s\n", *userID)

	// Sample order data
	orderData := map[string]interface{}{
		"order_request": map[string]interface{}{
			"order_id":   *orderID,
			"user_id":    *userID,
			"product_id": "PROD-456",
			"quantity":   2,
			"amount":     199.99,
			"shipping_address": map[string]string{
				"street":  "123 Main St",
				"city":    "New York",
				"country": "US",
				"zip":     "10001",
			},
		},
	}

	var result *workflow.ExecutionResult
	var err error

	switch *mode {
	case "programmatic":
		fmt.Println("🔧 Running in PROGRAMMATIC mode...")
		result, err = runProgrammaticMode(orderData)
	case "config":
		fmt.Println("📄 Running in CONFIG mode...")
		result, err = runConfigMode(orderData)
	case "autodiscovery":
		fmt.Println("🔍 Running in AUTO-DISCOVERY mode...")
		result, err = runAutoDiscoveryMode(orderData)
	default:
		log.Fatalf("❌ Invalid mode: %s. Use programmatic, config, or autodiscovery", *mode)
	}

	if err != nil {
		log.Fatalf("❌ Microservices orchestration execution failed: %v", err)
	}

	if result != nil {
		fmt.Printf("✅ Order processing completed successfully!\n")
		fmt.Printf("📊 Execution Summary:\n")
		fmt.Printf("   - Success: %t\n", result.Success)
		fmt.Printf("   - Duration: %dms\n", result.Duration)

		if result.Error != nil {
			fmt.Printf("   - Error: %s\n", result.Error.Error())
		}
	}
}

func runProgrammaticMode(data map[string]interface{}) (*workflow.ExecutionResult, error) {
	return programmatic.ExecuteWorkflow(data)
}

func runConfigMode(data map[string]interface{}) (*workflow.ExecutionResult, error) {
	// Load workflow from config file
	wm := workflow.NewWorkflowManager()

	// Register tasks first
	wm.RegisterTask("validateUser", tasks.ValidateUser)
	wm.RegisterTask("checkInventory", tasks.CheckInventory)
	wm.RegisterTask("reserveInventory", tasks.ReserveInventory)
	wm.RegisterTask("processPayment", tasks.ProcessPayment)
	wm.RegisterTask("createShipment", tasks.CreateShipment)
	wm.RegisterTask("sendNotification", tasks.SendNotification)

	// Load workflow from config
	err := wm.LoadFromConfig("./config/workflow.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Execute workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "validateUser", data)
}

func runAutoDiscoveryMode(data map[string]interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔍 Running in AUTO-DISCOVERY mode...")

	wm := workflow.NewWorkflowManager()

	// First, register tasks manually to ensure they're available
	fmt.Println("📝 Registering tasks manually as fallback...")
	wm.RegisterTask("validateUser", tasks.ValidateUser)
	wm.RegisterTask("checkInventory", tasks.CheckInventory)
	wm.RegisterTask("reserveInventory", tasks.ReserveInventory)
	wm.RegisterTask("processPayment", tasks.ProcessPayment)
	wm.RegisterTask("createShipment", tasks.CreateShipment)
	wm.RegisterTask("sendNotification", tasks.SendNotification)
	fmt.Println("✅ Manual task registration completed")

	// Then try auto-discovery as an additional feature
	fmt.Println("🔍 Attempting auto-discovery...")
	autodiscoverer := workflow.NewAutoDiscoverer()

	// Scan the tasks package for functions with annotations
	tasksPath := "./tasks"
	err := autodiscoverer.ScanPackage(tasksPath)
	if err != nil {
		fmt.Printf("⚠️  Auto-discovery failed: %v, continuing with manually registered tasks\n", err)
	} else {
		discoveredTasks := autodiscoverer.GetDiscoveredFunctions()
		fmt.Printf("   Found %d tasks through auto-discovery\n", len(discoveredTasks))

		// Register discovered tasks in the manager
		err = autodiscoverer.RegisterInManager(wm)
		if err != nil {
			fmt.Printf("⚠️  Failed to register auto-discovered tasks: %v\n", err)
		} else {
			fmt.Printf("   📝 Successfully registered auto-discovered tasks\n")
		}
	}

	// Load workflow from config
	err = wm.LoadFromConfig("./config/workflow.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Execute workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "validateUser", data)
}
