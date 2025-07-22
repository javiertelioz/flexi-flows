package main

import (
	"context"
	"github.com/javiertelioz/flexi-flows/examples/basic_workflow/tasks"
	"testing"
)

func TestBasicWorkflowTasks(t *testing.T) {
	ctx := context.Background()

	// Test ValidateData
	inputData := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30.0,
		},
	}

	result, err := tasks.ValidateData(ctx, inputData)
	if err != nil {
		t.Fatalf("ValidateData failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["validated"].(bool) {
		t.Error("Data should be validated")
	}

	// Test TransformData
	result, err = tasks.TransformData(ctx, result)
	if err != nil {
		t.Fatalf("TransformData failed: %v", err)
	}

	resultMap = result.(map[string]interface{})
	transformed := resultMap["transformed"].(map[string]interface{})
	if transformed["fullName"] != "JOHN DOE" {
		t.Error("Name should be transformed to uppercase")
	}

	t.Log("✅ Basic workflow tasks test passed")
}
