package programmatic

import (
	"context"
	"fmt"

	"github.com/javiertelioz/flexi-flows/examples/basic_workflow/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

// BuildWorkflowProgrammatically creates the workflow configuration programmatically
func BuildWorkflowProgrammatically() *workflow.WorkflowManager {
	wm := workflow.NewWorkflowManager()

	// Register tasks manually
	wm.RegisterTask("validateData", tasks.ValidateData)
	wm.RegisterTask("transformData", tasks.TransformData)
	wm.RegisterTask("saveData", tasks.SaveData)
	wm.RegisterTask("notifyUser", tasks.NotifyUser)

	// Create workflow configuration programmatically
	cfg := &config.WorkflowConfig{
		Name:        "basic_user_processing_workflow",
		Description: "Basic workflow built programmatically",
		Version:     "1.0.0",
		StartNode:   "validate",
		Nodes: []config.NodeConfig{
			{
				ID:       "validate",
				Type:     "task",
				Function: "validateData",
			},
			{
				ID:       "transform",
				Type:     "task",
				Function: "transformData",
			},
			{
				ID:       "save",
				Type:     "task",
				Function: "saveData",
			},
			{
				ID:       "notify",
				Type:     "task",
				Function: "notifyUser",
			},
		},
		Edges: []config.EdgeConfig{
			{From: "validate", To: "transform"},
			{From: "transform", To: "save"},
			{From: "save", To: "notify"},
		},
	}

	// Build workflow from config
	err := wm.BuildFromConfig(cfg)
	if err != nil {
		fmt.Printf("Failed to build workflow: %v\n", err)
		return nil
	}

	return wm
}

// ExecuteProgrammaticWorkflow runs the programmatically built workflow
func ExecuteProgrammaticWorkflow(inputData interface{}) (*workflow.ExecutionResult, error) {
	wm := BuildWorkflowProgrammatically()
	if wm == nil {
		return nil, fmt.Errorf("failed to build workflow")
	}

	ctx := context.Background()

	fmt.Println("🔧 Executing workflow built programmatically...")
	result, err := wm.ExecuteWithContext(ctx, "validate", inputData)

	if err != nil {
		return nil, fmt.Errorf("programmatic workflow execution failed: %w", err)
	}

	fmt.Println("✅ Programmatic workflow completed successfully!")
	return result, nil
}
