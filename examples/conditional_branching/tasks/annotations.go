package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// @workflow:task
// @name: classifyCustomer
// @description: Classifies customer type based on order data
func ClassifyCustomer(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	customerType, _ := dataMap["customer_type"].(string)
	orderValue, _ := dataMap["order_value"].(float64)
	customerID, _ := dataMap["customer_id"].(string)
	isFirstOrder, _ := dataMap["is_first_order"].(bool)

	// Classification logic
	classification := map[string]interface{}{
		"customer_id":    customerID,
		"order_value":    orderValue,
		"is_first_order": isFirstOrder,
		"original_type":  customerType,
	}

	switch {
	case customerType == "vip" || orderValue >= 1000:
		classification["category"] = "vip"
		classification["priority"] = "high"
		classification["branch"] = "vip_processing"
	case customerType == "regular" || (orderValue >= 100 && orderValue < 1000):
		classification["category"] = "regular"
		classification["priority"] = "normal"
		classification["branch"] = "regular_processing"
	case isFirstOrder || customerType == "new":
		classification["category"] = "new"
		classification["priority"] = "normal"
		classification["branch"] = "new_customer_processing"
	default:
		classification["category"] = "invalid"
		classification["priority"] = "low"
		classification["branch"] = "error_handling"
	}

	fmt.Printf("🏷️  Customer classified as: %s (Priority: %s)\n",
		classification["category"], classification["priority"])

	return classification, nil
}

// @workflow:task
// @name: processVIPOrder
// @description: Processes VIP customer orders with priority handling
func ProcessVIPOrder(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	customerID := dataMap["customer_id"].(string)
	orderValue := dataMap["order_value"].(float64)

	// VIP processing logic
	result := map[string]interface{}{
		"customer_id":      customerID,
		"order_value":      orderValue,
		"category":         dataMap["category"],
		"processing_type":  "vip_priority",
		"queue":            "priority_queue",
		"estimated_time":   "1-2 hours",
		"shipping_method":  "express",
		"discount_applied": orderValue * 0.1, // 10% VIP discount
		"personal_manager": true,
		"processed_at":     time.Now().Format(time.RFC3339),
	}

	fmt.Printf("⭐ VIP order processed for customer %s with express handling\n", customerID)
	return result, nil
}

// @workflow:task
// @name: processRegularOrder
// @description: Processes regular customer orders with standard handling
func ProcessRegularOrder(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	customerID := dataMap["customer_id"].(string)
	orderValue := dataMap["order_value"].(float64)

	// Regular processing logic
	result := map[string]interface{}{
		"customer_id":     customerID,
		"order_value":     orderValue,
		"category":        dataMap["category"],
		"processing_type": "standard",
		"queue":           "standard_queue",
		"estimated_time":  "3-5 hours",
		"shipping_method": "standard",
		"discount_applied": func() float64 {
			if orderValue >= 500 {
				return orderValue * 0.05 // 5% discount for orders over $500
			}
			return 0
		}(),
		"personal_manager": false,
		"processed_at":     time.Now().Format(time.RFC3339),
	}

	fmt.Printf("📦 Regular order processed for customer %s with standard handling\n", customerID)
	return result, nil
}

// @workflow:task
// @name: processNewCustomer
// @description: Processes new customer orders with verification and welcome
func ProcessNewCustomer(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	customerID := dataMap["customer_id"].(string)
	orderValue := dataMap["order_value"].(float64)

	// New customer processing with verification
	result := map[string]interface{}{
		"customer_id":           customerID,
		"order_value":           orderValue,
		"category":              dataMap["category"],
		"processing_type":       "new_customer_verification",
		"queue":                 "verification_queue",
		"estimated_time":        "4-6 hours",
		"shipping_method":       "standard",
		"verification_required": true,
		"welcome_package":       true,
		"first_order_discount":  orderValue * 0.15, // 15% welcome discount
		"onboarding_email":      true,
		"account_setup":         true,
		"processed_at":          time.Now().Format(time.RFC3339),
	}

	fmt.Printf("🎉 New customer order processed for %s with welcome benefits\n", customerID)
	return result, nil
}

// @workflow:task
// @name: handleInvalidOrder
// @description: Handles invalid orders with error processing
func HandleInvalidOrder(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	customerID := dataMap["customer_id"].(string)
	orderValue := dataMap["order_value"].(float64)

	// Error handling logic
	errors := []string{}
	if customerID == "" {
		errors = append(errors, "missing customer ID")
	}
	if orderValue <= 0 {
		errors = append(errors, "invalid order value")
	}

	result := map[string]interface{}{
		"customer_id":            customerID,
		"order_value":            orderValue,
		"category":               "invalid",
		"processing_type":        "error_handling",
		"status":                 "failed",
		"errors":                 errors,
		"requires_manual_review": true,
		"notification_sent":      true,
		"processed_at":           time.Now().Format(time.RFC3339),
	}

	fmt.Printf("❌ Invalid order handled for customer %s - Errors: %s\n",
		customerID, strings.Join(errors, ", "))
	return result, nil
}

// @workflow:task
// @name: finalizeOrder
// @description: Finalizes order processing regardless of branch taken
func FinalizeOrder(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	customerID := dataMap["customer_id"].(string)
	category := dataMap["category"].(string)
	processingType := dataMap["processing_type"].(string)

	// Final processing
	result := map[string]interface{}{
		"order_finalized":   true,
		"customer_id":       customerID,
		"category":          category,
		"processing_type":   processingType,
		"confirmation_sent": true,
		"tracking_number":   fmt.Sprintf("TRK-%s-%d", strings.ToUpper(category), time.Now().Unix()),
		"finalized_at":      time.Now().Format(time.RFC3339),
	}

	// Copy all previous data
	for key, value := range dataMap {
		if key != "order_finalized" && key != "confirmation_sent" && key != "tracking_number" && key != "finalized_at" {
			result[key] = value
		}
	}

	fmt.Printf("✅ Order finalized for customer %s (%s processing)\n", customerID, processingType)
	return result, nil
}
