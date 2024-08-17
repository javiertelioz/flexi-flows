package subdag_flow

import (
	"log"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"

	"github.com/javiertelioz/flexi-flows/examples/subdag_flow/uses_cases"
)

func SubDagFlow() {
	wm := workflow.NewWorkflowManager()

	// Register task & hooks
	wm.RegisterTask("SubTask1", uses_cases.SubTask1)
	wm.RegisterTask("SubTask2", uses_cases.SubTask2)
	wm.RegisterTask("MainTask", uses_cases.MainTask)

	// Load workflow configuration
	workflowConfig, err := config.LoadConfig("./examples/subdag_flow/config/workflow.json")
	if err != nil {
		log.Fatalf("Failed to load workflow configuration: %v", err)
	}

	// Load nodes and edges from configuration
	err = wm.LoadFromConfig(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to load workflow from configuration: %v", err)
	}

	// Execute workflow
	err = wm.Execute("mainTask", nil)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
}
