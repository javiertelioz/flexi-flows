package programmatic

import (
	"context"
	"fmt"

	"github.com/javiertelioz/flexi-flows/examples/data_processing_pipeline/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
	"github.com/javiertelioz/flexi-flows/pkg/workflow/config"
)

// ExecuteDataProcessingPipeline creates and executes the data processing pipeline programmatically
func ExecuteDataProcessingPipeline(inputData interface{}) (*workflow.ExecutionResult, error) {
	fmt.Println("🔧 Building data processing pipeline programmatically...")

	// Create workflow manager
	wm := workflow.NewWorkflowManager()

	// Register all tasks
	wm.RegisterTask("extractData", tasks.ExtractData)
	wm.RegisterTask("validateData", tasks.ValidateData)
	wm.RegisterTask("transformData", tasks.TransformData)
	wm.RegisterTask("enrichData", tasks.EnrichData)
	wm.RegisterTask("filterData", tasks.FilterData)
	wm.RegisterTask("aggregateData", tasks.AggregateData)
	wm.RegisterTask("saveData", tasks.SaveData)

	// Create workflow configuration programmatically
	cfg := buildDataPipelineConfig()

	// Build from configuration
	err := wm.BuildFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build workflow: %w", err)
	}

	// Execute the workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "extractData", inputData)
}

func buildDataPipelineConfig() *config.WorkflowConfig {
	cfg := &config.WorkflowConfig{
		Name:        "DataProcessingPipelineProgrammatic",
		Description: "Complete data processing pipeline built programmatically",
		Version:     "1.0.0",
		StartNode:   "extractData",
		Settings: config.WorkflowSettings{
			MaxRetries:  3,
			Timeout:     "60s",
			EnableDebug: true,
		},
		Variables: map[string]interface{}{
			"batch_size":           10,
			"validation_threshold": 0.8,
			"min_user_score":       3.0,
		},
		Nodes: []config.NodeConfig{
			{
				ID:          "extractData",
				Type:        "task",
				Name:        "Extract Data",
				Description: "Extract data from various sources",
				Function:    "extractData",
				Next:        []string{"validateData"},
			},
			{
				ID:          "validateData",
				Type:        "task",
				Name:        "Validate Data",
				Description: "Validate extracted data for completeness",
				Function:    "validateData",
				Next:        []string{"transformData"},
			},
			{
				ID:          "transformData",
				Type:        "task",
				Name:        "Transform Data",
				Description: "Transform and normalize data fields",
				Function:    "transformData",
				Next:        []string{"enrichData"},
			},
			{
				ID:          "enrichData",
				Type:        "task",
				Name:        "Enrich Data",
				Description: "Enrich data with additional information",
				Function:    "enrichData",
				Next:        []string{"filterData"},
			},
			{
				ID:          "filterData",
				Type:        "task",
				Name:        "Filter Data",
				Description: "Filter data based on business rules",
				Function:    "filterData",
				Next:        []string{"aggregateData"},
			},
			{
				ID:          "aggregateData",
				Type:        "task",
				Name:        "Aggregate Data",
				Description: "Aggregate and summarize processed data",
				Function:    "aggregateData",
				Next:        []string{"saveData"},
			},
			{
				ID:          "saveData",
				Type:        "task",
				Name:        "Save Data",
				Description: "Save processed data to storage",
				Function:    "saveData",
			},
		},
	}

	fmt.Println("   ✅ Programmatic pipeline configuration built successfully")
	fmt.Println("   📊 Pipeline structure:")
	fmt.Println("      extractData → validateData → transformData → enrichData")
	fmt.Println("      → filterData → aggregateData → saveData")

	return cfg
}
