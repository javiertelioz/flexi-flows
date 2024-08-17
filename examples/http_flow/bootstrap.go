package http_flow

import (
	"github.com/javiertelioz/flexi-flows/examples/http_flow/uses_cases"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
	"log"
)

func HttpFlow() {
	wm := workflow.NewWorkflowManager()

	// Register storage
	// stateStore := storage.NewJSONStateStore("flow.json")
	// wm.RegisterStateStore(stateStore)

	// Register task & hooks
	wm.RegisterTask("getUserFunc", use_cases.GetUserFunc)
	wm.RegisterTask("isPrimeFunc", use_cases.IsPrimeFunc)
	wm.RegisterTask("squareFunc", use_cases.SquareFunc)
	wm.RegisterTask("sumFunc", use_cases.SumFunc)

	wm.RegisterHook("beforeIsPrime", use_cases.BeforeIsPrime)
	wm.RegisterHook("afterIsPrime", use_cases.AfterIsPrime)
	wm.RegisterHook("beforeSquare", use_cases.BeforeSquare)
	wm.RegisterHook("afterSquare", use_cases.AfterSquare)
	wm.RegisterHook("beforeSum", use_cases.BeforeSum)
	wm.RegisterHook("afterSum", use_cases.AfterSum)

	workflowConfig, err := config.LoadConfig("config/workflow.json")
	if err != nil {
		log.Fatalf("Failed to load workflow configuration: %v", err)
	}

	err = wm.LoadFromConfig(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to load workflow from configuration: %v", err)
	}

	err = wm.Execute("parallel", 5)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
}
