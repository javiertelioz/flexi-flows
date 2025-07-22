package programmatic

import (
	"context"
	"fmt"

	"github.com/javiertelioz/flexi-flows/examples/microservices_orchestration/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

// ExecuteWorkflow creates and executes the microservices orchestration workflow programmatically
func ExecuteWorkflow(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔧 Building microservices orchestration workflow programmatically...")

	// Create workflow manager
	wm := workflow.NewWorkflowManager()

	// Register all tasks
	wm.RegisterTask("validateUser", tasks.ValidateUser)
	wm.RegisterTask("checkInventory", tasks.CheckInventory)
	wm.RegisterTask("reserveInventory", tasks.ReserveInventory)
	wm.RegisterTask("processPayment", tasks.ProcessPayment)
	wm.RegisterTask("createShipment", tasks.CreateShipment)
	wm.RegisterTask("sendNotification", tasks.SendNotification)

	// Create workflow configuration programmatically
	cfg := buildMicroservicesOrchestrationConfig()

	// Build from configuration
	err := wm.BuildFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build workflow: %w", err)
	}

	// Execute the workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "validateUser", inputData)
}

// buildMicroservicesOrchestrationConfig creates the workflow configuration programmatically
func buildMicroservicesOrchestrationConfig() *config.WorkflowConfig {
	return &config.WorkflowConfig{
		Name:        "MicroservicesOrchestration",
		Description: "Orchestrates order processing across multiple microservices",
		Version:     "1.0.0",
		StartNode:   "validateUser",
		Settings: config.WorkflowSettings{
			MaxRetries:  3,
			Timeout:     "30s",
			EnableDebug: true,
		},
		Variables: map[string]interface{}{
			"payment_timeout":    "5s",
			"inventory_timeout":  "3s",
			"notification_retry": 2,
		},
		Nodes: []config.NodeConfig{
			{
				ID:          "validateUser",
				Type:        "task",
				Name:        "Validate User",
				Description: "Validate user credentials and permissions",
				Function:    "validateUser",
				Next:        []string{"checkInventory"},
			},
			{
				ID:          "checkInventory",
				Type:        "task",
				Name:        "Check Inventory",
				Description: "Check product availability in inventory service",
				Function:    "checkInventory",
				Next:        []string{"reserveInventory"},
			},
			{
				ID:          "reserveInventory",
				Type:        "task",
				Name:        "Reserve Inventory",
				Description: "Reserve items in inventory for the order",
				Function:    "reserveInventory",
				Next:        []string{"processPayment"},
			},
			{
				ID:          "processPayment",
				Type:        "task",
				Name:        "Process Payment",
				Description: "Process payment through payment service",
				Function:    "processPayment",
				Next:        []string{"createShipment"},
			},
			{
				ID:          "createShipment",
				Type:        "task",
				Name:        "Create Shipment",
				Description: "Create shipment in logistics service",
				Function:    "createShipment",
				Next:        []string{"sendNotification"},
			},
			{
				ID:          "sendNotification",
				Type:        "task",
				Name:        "Send Notification",
				Description: "Send confirmation notification to customer",
				Function:    "sendNotification",
				Next:        []string{},
			},
		},
	}
}
