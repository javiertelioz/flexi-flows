package programmatic

import (
	"context"
	"fmt"

	"github.com/javiertelioz/flexi-flows/examples/data_processing_pipeline/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

// BuildDataPipelineProgrammatically creates the data processing pipeline entirely through code
func BuildDataPipelineProgrammatically() *workflow.WorkflowManager {
	wm := workflow.NewWorkflowManager()

	// Register tasks manually
	wm.RegisterTask("fetchUsers", tasks.FetchUsers)
	wm.RegisterTask("validateUsers", tasks.ValidateUsers)
	wm.RegisterTask("transformUsers", tasks.TransformUsers)
	wm.RegisterTask("saveUsers", tasks.SaveUsers)
	wm.RegisterTask("generateReport", tasks.GenerateReport)

	// Create workflow configuration programmatically
	cfg := &config.WorkflowConfig{
		Name:        "data_processing_pipeline_programmatic",
		Description: "Data processing pipeline built programmatically",
		Version:     "1.0.0",
		StartNode:   "fetch",
		Nodes: []config.NodeConfig{
			{
				ID:       "fetch",
				Type:     "task",
				Function: "fetchUsers",
			},
			{
				ID:       "validate",
				Type:     "task",
				Function: "validateUsers",
			},
			{
				ID:       "transform",
				Type:     "task",
				Function: "transformUsers",
			},
			{
				ID:       "save",
				Type:     "task",
				Function: "saveUsers",
			},
			{
				ID:       "report",
				Type:     "task",
				Function: "generateReport",
			},
		},
		Edges: []config.EdgeConfig{
			{From: "fetch", To: "validate"},
			{From: "validate", To: "transform"},
			{From: "transform", To: "save"},
			{From: "save", To: "report"},
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

// ExecuteDataPipelineWorkflow runs the data processing pipeline
func ExecuteDataPipelineWorkflow(inputData interface{}) (*workflow.ExecutionResult, error) {
	wm := BuildDataPipelineProgrammatically()
	if wm == nil {
		return nil, fmt.Errorf("failed to build workflow")
	}

	ctx := context.Background()

	fmt.Println("🔧 Executing data processing pipeline built programmatically...")
	result, err := wm.ExecuteWithContext(ctx, "fetch", inputData)

	if err != nil {
		return nil, fmt.Errorf("data pipeline execution failed: %w", err)
	}

	fmt.Println("✅ Data processing pipeline completed successfully!")
	return result, nil
}
