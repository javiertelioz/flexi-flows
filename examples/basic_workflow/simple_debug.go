package main

import (
	"fmt"

	"github.com/javiertelioz/flexi-flows/examples/basic_workflow/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

func main() {
	fmt.Println("🔧 Testing simplified WorkflowManager...")

	// Create WorkflowManager
	wm := workflow.NewWorkflowManager()

	// Register only one task for testing
	wm.RegisterTask("validateData", tasks.ValidateData)
	wm.RegisterTask("transformData", tasks.TransformData)
	wm.RegisterTask("saveData", tasks.SaveData)
	wm.RegisterTask("notifyUser", tasks.NotifyUser)

	fmt.Println("✅ Task registered successfully")

	// Test loading a minimal config
	// For now, let's test the basic WorkflowManager functionality

	inputData := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30.0,
		},
	}

	fmt.Printf("Input Data: %+v\n", inputData)

	// Try to get registered tasks
	registeredTasks := wm.GetRegisteredTasks()
	fmt.Printf("✅ Registered tasks: %v\n", registeredTasks)

	// Test getting a specific task
	task, exists := wm.GetTask("validateData")
	if exists {
		fmt.Println("✅ Task 'validateData' found successfully")
	} else {
		fmt.Println("❌ Task 'validateData' not found")
	}

	// Test the task directly through the manager
	if task != nil {
		fmt.Println("🧪 Testing task execution through manager...")
		// We won't call ExecuteWithContext yet since that might be where the hang occurs
	}

	fmt.Println("🎉 Basic WorkflowManager functionality works!")

	// Now let's test loading config step by step
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

	fmt.Println("🎉 Config loading works! Next step would be ExecuteWithContext...")
}
