package tasks

import (
	"context"
	"fmt"
	"time"
)

// @workflow:task
// @name: classifyOrder
// @description: Classifies order based on customer type and order value
func ClassifyOrder(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🔍 Classifying order...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	order, ok := dataMap["order"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("order data not found")
	}

	customerType, _ := order["customer_type"].(string)
	orderValue, _ := order["order_value"].(float64)

	// Classification logic
	var category string
	var priority int

	switch customerType {
	case "vip":
		category = "vip"
		priority = 1
	case "regular":
		if orderValue >= 100 {
			category = "regular_high"
			priority = 2
		} else {
			category = "regular_low"
			priority = 3
		}
	case "new":
		category = "new_customer"
		priority = 4
	default:
		category = "invalid"
		priority = 5
	}

	result := map[string]interface{}{
		"original_order":      order,
		"category":            category,
		"priority":            priority,
		"customer_type":       customerType,
		"order_value":         orderValue,
		"classification_time": time.Now().Unix(),
	}

	fmt.Printf("   ✅ Order classified as: %s (priority: %d)\n", category, priority)
	return result, nil
}

// @workflow:task
// @name: processVIPOrder
// @description: Process VIP customer orders with special handling
func ProcessVIPOrder(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("👑 Processing VIP order...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	// Simulate VIP processing
	time.Sleep(100 * time.Millisecond)

	result := map[string]interface{}{
		"original_data":      dataMap,
		"processing_type":    "vip",
		"discount_applied":   "15%",
		"expedited_shipping": true,
		"priority_support":   true,
		"loyalty_points":     500,
		"processed_at":       time.Now().Unix(),
		"status":             "vip_processed",
	}

	fmt.Println("   ✅ VIP order processed with premium benefits")
	return result, nil
}

// @workflow:task
// @name: processRegularOrder
// @description: Process regular customer orders
func ProcessRegularOrder(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🛍️ Processing regular order...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	category, _ := dataMap["category"].(string)
	orderValue, _ := dataMap["order_value"].(float64)

	// Simulate regular processing
	time.Sleep(50 * time.Millisecond)

	var discount string
	var loyaltyPoints int
	if category == "regular_high" {
		discount = "5%"
		loyaltyPoints = 100
	} else {
		discount = "0%"
		loyaltyPoints = 50
	}

	result := map[string]interface{}{
		"original_data":      dataMap,
		"processing_type":    "regular",
		"discount_applied":   discount,
		"expedited_shipping": orderValue >= 200,
		"loyalty_points":     loyaltyPoints,
		"processed_at":       time.Now().Unix(),
		"status":             "regular_processed",
	}

	fmt.Printf("   ✅ Regular order processed with %s discount\n", discount)
	return result, nil
}

// @workflow:task
// @name: processNewCustomer
// @description: Process new customer orders with welcome benefits
func ProcessNewCustomer(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🆕 Processing new customer order...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	// Simulate new customer processing
	time.Sleep(75 * time.Millisecond)

	result := map[string]interface{}{
		"original_data":        dataMap,
		"processing_type":      "new_customer",
		"welcome_discount":     "10%",
		"expedited_shipping":   false,
		"welcome_bonus_points": 200,
		"email_welcome_sent":   true,
		"processed_at":         time.Now().Unix(),
		"status":               "new_customer_processed",
	}

	fmt.Println("   ✅ New customer order processed with welcome benefits")
	return result, nil
}

// @workflow:task
// @name: handleInvalidOrder
// @description: Handle invalid orders and customer types
func HandleInvalidOrder(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("❌ Handling invalid order...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	// Simulate error handling
	time.Sleep(25 * time.Millisecond)

	result := map[string]interface{}{
		"original_data":   dataMap,
		"processing_type": "invalid",
		"error_handled":   true,
		"support_ticket":  "TKT_" + fmt.Sprintf("%d", time.Now().Unix()),
		"escalated":       true,
		"processed_at":    time.Now().Unix(),
		"status":          "invalid_handled",
	}

	fmt.Println("   ✅ Invalid order handled and escalated to support")
	return result, nil
}

// @workflow:task
// @name: finalizeOrder
// @description: Finalize order processing and prepare for fulfillment
func FinalizeOrder(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🎯 Finalizing order...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	// Simulate finalization
	time.Sleep(30 * time.Millisecond)

	processingType, _ := dataMap["processing_type"].(string)
	status, _ := dataMap["status"].(string)

	result := map[string]interface{}{
		"original_data":      dataMap,
		"finalized_at":       time.Now().Unix(),
		"processing_type":    processingType,
		"previous_status":    status,
		"final_status":       "finalized",
		"ready_for_shipment": true,
		"tracking_number":    "TRK_" + fmt.Sprintf("%d", time.Now().Unix()),
		"estimated_delivery": time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
	}

	fmt.Printf("   ✅ Order finalized and ready for shipment (Type: %s)\n", processingType)
	return result, nil
}

// Conditional functions for workflow branching
// @workflow:conditional
// @name: isVIPCustomer
// @description: Check if customer is VIP
func IsVIPCustomer(data interface{}) bool {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	category, exists := dataMap["category"].(string)
	if !exists {
		return false
	}

	return category == "vip"
}

// @workflow:conditional
// @name: isRegularCustomer
// @description: Check if customer is regular
func IsRegularCustomer(data interface{}) bool {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	category, exists := dataMap["category"].(string)
	if !exists {
		return false
	}

	return category == "regular_high" || category == "regular_low"
}

// @workflow:conditional
// @name: isNewCustomer
// @description: Check if customer is new
func IsNewCustomer(data interface{}) bool {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	category, exists := dataMap["category"].(string)
	if !exists {
		return false
	}

	return category == "new_customer"
}
