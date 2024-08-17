package task_flow

import (
	"log"

	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"

	"github.com/javiertelioz/flexi-flows/examples/task_flow/uses_cases"
)

func TaskFlow() {
	wm := workflow.NewWorkflowManager()

	// Register task & hooks
	wm.RegisterTask("Task1", uses_cases.Task1)
	wm.RegisterTask("Task2", uses_cases.Task2)
	wm.RegisterTask("Task3", uses_cases.Task3)

	// Load workflow configuration
	workflowConfig, err := config.LoadConfig("./examples/task_flow/config/workflow.json")
	if err != nil {
		log.Fatalf("Failed to load workflow configuration: %v", err)
	}

	// Load nodes and edges from configuration
	err = wm.LoadFromConfig(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to load workflow from configuration: %v", err)
	}

	// Execute workflow
	err = wm.Execute("task1", "nil")
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
}
