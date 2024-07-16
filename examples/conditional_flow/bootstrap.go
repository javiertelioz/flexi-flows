package conditional_flow

import (
	"log"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/javiertelioz/go-flows/pkg/workflow/config"

	"github.com/javiertelioz/go-flows/examples/conditional_flow/uses_cases"
)

func ConditionalFlow() {
	wm := workflow.NewWorkflowManager()

	// Register task & hooks
	wm.RegisterTask("ConditionFunc", uses_cases.ConditionFunc)
	wm.RegisterTask("TrueBranch", uses_cases.TrueBranch)
	wm.RegisterTask("FalseBranch", uses_cases.FalseBranch)

	// Load workflow configuration
	workflowConfig, err := config.LoadConfig("./examples/conditional_flow/config/workflow.json")
	if err != nil {
		log.Fatalf("Failed to load workflow configuration: %v", err)
	}

	// Load nodes and edges from configuration
	err = wm.LoadFromConfig(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to load workflow from configuration: %v", err)
	}

	// Execute workflow
	err = wm.Execute("condition", "some initial data")
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
}
