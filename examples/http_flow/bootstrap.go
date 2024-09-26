package http_flow

import (
	"log"

	"github.com/javiertelioz/flexi-flows/examples/http_flow/uses_cases"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

type User struct {
	ID    int
	Name  string
	Email string
}

func HttpFlow() {
	wm := workflow.NewWorkflowManager[any]()

	// Register tasks & hooks
	wm.RegisterTask("getUserFunc", use_cases.GetUserFunc)
	wm.RegisterTask("isPrimeFunc", use_cases.IsPrimeFunc)
	wm.RegisterTask("squareFunc", use_cases.SquareFunc)
	wm.RegisterTask("sumFunc", use_cases.SumFunc)
	wm.RegisterTask("iterateFunc", use_cases.IterateFunc)

	wm.RegisterHook("beforeIsPrime", use_cases.BeforeIsPrime)
	wm.RegisterHook("afterIsPrime", use_cases.AfterIsPrime)
	wm.RegisterHook("beforeSquare", use_cases.BeforeSquare)
	wm.RegisterHook("afterSquare", use_cases.AfterSquare)
	wm.RegisterHook("beforeSum", use_cases.BeforeSum)
	wm.RegisterHook("afterSum", use_cases.AfterSum)

	workflowConfig, err := config.LoadConfig("./examples/http_flow/config/workflow.json")

	if err != nil {
		log.Fatalf("Failed to load workflow configuration: %v", err)
	}

	err = wm.LoadFromConfig(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to load workflow from configuration: %v", err)
	}

	//items := []any{"1", "3", "5"}
	/*items := []any{
		User{ID: 1, Name: "Javier", Email: "javier@example.com"},
		User{ID: 2, Name: "Ana", Email: "ana@example.com"},
		User{ID: 3, Name: "Carlos", Email: "carlos@example.com"},
	}*/
	// Execute the workflow starting from 'foreachNode'
	//result, err := wm.Execute("foreachNode", items)
	result, err := wm.Execute("parallel", 5)

	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	log.Printf("Workflow completed with result: %v", result)
}
