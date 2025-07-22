package main

import (
	"context"
	"fmt"
	"log"

	"github.com/javiertelioz/flexi-flows/examples/conditional_branching/tasks"
	"github.com/javiertelioz/flexi-flows/pkg/workflow"
)

func main() {
	fmt.Println("🧪 Testing conditional branching tasks individually...")

	// Test the classifyOrder task directly
	inputData := map[string]interface{}{
		"order": map[string]interface{}{
			"id":            "ORD_vip_001",
			"customer_id":   "CUST_vip_123",
			"customer_type": "vip",
			"order_value":   500.0,
		},
	}

	ctx := context.Background()

	// Test classification
	result, err := tasks.ClassifyOrder(ctx, inputData)
	if err != nil {
		log.Fatalf("Classification failed: %v", err)
	}

	fmt.Printf("✅ Classification result: %+v\n", result)

	// Test VIP processing
	vipResult, err := tasks.ProcessVIPOrder(ctx, result)
	if err != nil {
		log.Fatalf("VIP processing failed: %v", err)
	}

	fmt.Printf("✅ VIP processing result: %+v\n", vipResult)

	// Test finalization
	finalResult, err := tasks.FinalizeOrder(ctx, vipResult)
	if err != nil {
		log.Fatalf("Finalization failed: %v", err)
	}

	fmt.Printf("✅ Final result: %+v\n", finalResult)
	fmt.Println("🎉 Individual task tests completed successfully!")
}
