package programmatic

import (
	"context"
	"fmt"

	"github.com/javiertelioz/flexi-flows/examples/conditional_branching/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

// ExecuteConditionalBranchingWorkflow demonstrates programmatic workflow creation with conditional branching
func ExecuteConditionalBranchingWorkflow(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔧 Building workflow programmatically...")

	// Create workflow manager
	wm := workflow.NewWorkflowManager()

	// Register all tasks
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

	// Create workflow configuration programmatically
	cfg := buildConditionalWorkflowConfig()

	// Build from configuration
	err := wm.BuildFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build workflow: %w", err)
	}

	// Execute the workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "classifyOrder", inputData)
}

func buildConditionalWorkflowConfig() *config.WorkflowConfig {
	cfg := &config.WorkflowConfig{
		Name:        "ConditionalBranchingWorkflowProgrammatic",
		Description: "Order processing workflow with conditional branching built programmatically",
		Version:     "1.0.0",
		StartNode:   "classifyOrder",
		Settings: config.WorkflowSettings{
			MaxRetries:  3,
			Timeout:     "30s",
			EnableDebug: true,
		},
		Variables: map[string]interface{}{
			"vip_discount":       0.15,
			"regular_discount":   0.05,
			"new_customer_bonus": 200,
		},
		Nodes: []config.NodeConfig{
			{
				ID:          "classifyOrder",
				Type:        "task",
				Name:        "Classify Order",
				Description: "Classify order based on customer type and order value",
				Function:    "classifyOrder",
				Next:        []string{"vipCondition"},
			},
			{
				ID:          "vipCondition",
				Type:        "conditional",
				Name:        "VIP Customer Check",
				Description: "Check if customer is VIP",
				Condition:   "isVIPCustomer",
				TruePath:    []string{"processVIPBranch"},
				FalsePath:   []string{"checkRegularCustomer"},
			},
			{
				ID:          "checkRegularCustomer",
				Type:        "conditional",
				Name:        "Regular Customer Check",
				Description: "Check if customer is regular",
				Condition:   "isRegularCustomer",
				TruePath:    []string{"processRegularBranch"},
				FalsePath:   []string{"checkNewCustomer"},
			},
			{
				ID:          "checkNewCustomer",
				Type:        "conditional",
				Name:        "New Customer Check",
				Description: "Check if customer is new",
				Condition:   "isNewCustomer",
				TruePath:    []string{"processNewCustomerBranch"},
				FalsePath:   []string{"handleInvalidBranch"},
			},
			{
				ID:          "processVIPBranch",
				Type:        "task",
				Name:        "Process VIP Order",
				Description: "Process VIP customer order with premium benefits",
				Function:    "processVIPOrder",
				Next:        []string{"finalizeOrder"},
			},
			{
				ID:          "processRegularBranch",
				Type:        "task",
				Name:        "Process Regular Order",
				Description: "Process regular customer order",
				Function:    "processRegularOrder",
				Next:        []string{"finalizeOrder"},
			},
			{
				ID:          "processNewCustomerBranch",
				Type:        "task",
				Name:        "Process New Customer",
				Description: "Process new customer order with welcome benefits",
				Function:    "processNewCustomer",
				Next:        []string{"finalizeOrder"},
			},
			{
				ID:          "handleInvalidBranch",
				Type:        "task",
				Name:        "Handle Invalid Order",
				Description: "Handle invalid orders and customer types",
				Function:    "handleInvalidOrder",
				Next:        []string{"finalizeOrder"},
			},
			{
				ID:          "finalizeOrder",
				Type:        "task",
				Name:        "Finalize Order",
				Description: "Finalize order processing and prepare for fulfillment",
				Function:    "finalizeOrder",
			},
		},
	}

	fmt.Println("   ✅ Programmatic workflow configuration built successfully")
	fmt.Println("   📊 Workflow structure:")
	fmt.Println("      classifyOrder → vipCondition → [processVIPBranch | checkRegularCustomer]")
	fmt.Println("      checkRegularCustomer → [processRegularBranch | checkNewCustomer]")
	fmt.Println("      checkNewCustomer → [processNewCustomerBranch | handleInvalidBranch]")
	fmt.Println("      All branches → finalizeOrder")

	return cfg
}
