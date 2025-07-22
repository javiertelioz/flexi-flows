package programmatic

import (
	"context"

	"github.com/javiertelioz/flexi-flows/examples/parallel_processing/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

// ExecuteParallelWeatherWorkflow creates and executes a parallel weather processing workflow programmatically
func ExecuteParallelWeatherWorkflow(inputData interface{}) (*workflow.ExecutionResult, error) {
	// Create workflow manager
	wm := workflow.NewWorkflowManager()

	// Register tasks
	wm.RegisterTask("fetchWeatherData", tasks.FetchWeatherData)
	wm.RegisterTask("aggregateWeatherData", tasks.AggregateWeatherData)

	// Create nodes programmatically
	fetchNode := wm.CreateTaskNode("fetch_weather", "fetchWeatherData")
	aggregateNode := wm.CreateTaskNode("aggregate_weather", "aggregateWeatherData")

	// Connect nodes
	fetchNode.Next = []string{"aggregate_weather"}

	// Add nodes to workflow
	wm.AddNode("fetch_weather", fetchNode)
	wm.AddNode("aggregate_weather", aggregateNode)

	// Set start node
	wm.SetStartNode("fetch_weather")

	// Execute workflow
	ctx := context.Background()
	return wm.ExecuteWithContext(ctx, "fetch_weather", inputData)
}
