package main

import (
	"context"
	"fmt"

	"github.com/javiertelioz/flexi-flows/examples/data_processing_pipeline/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

func main() {
	fmt.Println("🔧 Testing data_processing_pipeline WorkflowManager...")

	// Create WorkflowManager
	wm := workflow.NewWorkflowManager()

	// Register tasks
	wm.RegisterTask("fetchUsers", tasks.FetchUsers)
	wm.RegisterTask("validateUsers", tasks.ValidateUsers)
	wm.RegisterTask("transformUsers", tasks.TransformUsers)
	wm.RegisterTask("saveUsers", tasks.SaveUsers)
	wm.RegisterTask("generateReport", tasks.GenerateReport)

	fmt.Println("✅ Tasks registered successfully")

	// Try to get registered tasks
	registeredTasks := wm.GetRegisteredTasks()
	fmt.Printf("✅ Registered tasks: %v\n", registeredTasks)

	// Test loading config
	fmt.Println("\n🔧 Testing config loading...")

	err := wm.LoadFromConfig("./config/workflow.json")
	if err != nil {
		fmt.Printf("❌ LoadFromConfig failed: %v\n", err)
		return
	}

	fmt.Println("✅ Config loaded successfully!")

	// Let's see what's in the graph
	graph := wm.GetGraph()
	fmt.Printf("📊 Graph nodes count: %d\n", len(graph.Nodes))
	fmt.Printf("📊 Graph edges count: %d\n", len(graph.Edges))

	fmt.Println("🎉 Config loading works!")

	// Test individual tasks
	fmt.Println("\n🧪 Testing individual tasks...")

	inputData := map[string]interface{}{
		"source":     "mock",
		"batch_size": 2,
	}

	// Test FetchUsers task
	fmt.Println("Testing FetchUsers...")
	result, err := tasks.FetchUsers(context.Background(), inputData)
	if err != nil {
		fmt.Printf("❌ FetchUsers failed: %v\n", err)
		return
	}
	fmt.Printf("✅ FetchUsers result keys: %v\n", getMapKeys(result))

	fmt.Println("🎉 Individual tasks work!")
}

func getMapKeys(data interface{}) []string {
	if dataMap, ok := data.(map[string]interface{}); ok {
		keys := make([]string, 0, len(dataMap))
		for key := range dataMap {
			keys = append(keys, key)
		}
		return keys
	}
	return []string{}
}
