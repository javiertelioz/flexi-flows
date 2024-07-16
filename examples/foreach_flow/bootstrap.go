package foreach_flow

import (
	"log"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/javiertelioz/go-flows/pkg/workflow/config"

	"github.com/javiertelioz/go-flows/examples/foreach_flow/uses_cases"
)

func ForeachFlow() {
	wm := workflow.NewWorkflowManager()

	// Register task & hooks
	wm.RegisterTask("IterateFunc", uses_cases.IterateFunc)

	// Load workflow configuration
	workflowConfig, err := config.LoadConfig("./examples/foreach_flow/config/workflow.json")
	if err != nil {
		log.Fatalf("Failed to load workflow configuration: %v", err)
	}

	// Load nodes and edges from configuration
	err = wm.LoadFromConfig(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to load workflow from configuration: %v", err)
	}

	// Execute workflow
	err = wm.Execute("foreach1", nil)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
}
