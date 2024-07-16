package branch_flow

import (
	"log"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/javiertelioz/go-flows/pkg/workflow/config"

	"github.com/javiertelioz/go-flows/examples/branch_flow/uses_cases"
)

func BranchFlow() {
	wm := workflow.NewWorkflowManager()

	// Register task & hooks
	wm.RegisterTask("Branch1", uses_cases.Branch1)
	wm.RegisterTask("Branch2", uses_cases.Branch2)
	wm.RegisterTask("Branch3", uses_cases.Branch3)

	// Load workflow configuration
	workflowConfig, err := config.LoadConfig("./examples/branch_flow/config/workflow.json")
	if err != nil {
		log.Fatalf("Failed to load workflow configuration: %v", err)
	}

	// Load nodes and edges from configuration
	err = wm.LoadFromConfig(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to load workflow from configuration: %v", err)
	}

	// Execute workflow
	err = wm.Execute("branch1", nil)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
}
