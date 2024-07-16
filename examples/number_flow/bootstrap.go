package number_flow

import (
	"log"

	"github.com/javiertelioz/go-flows/pkg/workflow"
	"github.com/javiertelioz/go-flows/pkg/workflow/config"

	"github.com/javiertelioz/go-flows/examples/number_flow/uses_cases"
)

func NumberFlow() {
	wm := workflow.NewWorkflowManager()

	// Register task & hooks
	wm.RegisterTask("Task1", uses_cases.Task1)
	wm.RegisterTask("Task2", uses_cases.Task2)
	wm.RegisterTask("SubTask1", uses_cases.SubTask1)
	wm.RegisterTask("SubTask2", uses_cases.SubTask2)
	wm.RegisterTask("IterateFunc", uses_cases.IterateFunc)
	wm.RegisterTask("ConditionFunc", uses_cases.ConditionFunc)

	wm.RegisterHook("BeforeTask2", uses_cases.BeforeTask2)
	wm.RegisterHook("AfterTask1", uses_cases.AfterTask1)
	wm.RegisterHook("BeforeSubTask1", uses_cases.BeforeSubTask1)
	wm.RegisterHook("AfterSubTask1", uses_cases.AfterSubTask1)

	// Load workflow configuration
	workflowConfig, err := config.LoadConfig("./examples/number_flow/config/workflow.json")
	if err != nil {
		log.Fatalf("Failed to load workflow configuration: %v", err)
	}

	// Load nodes and edges from configuration
	err = wm.LoadFromConfig(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to load workflow from configuration: %v", err)
	}

	// Execute workflow
	err = wm.Execute("task1", "start")
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
}
