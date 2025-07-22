package tasks

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// @workflow:task
// @name: validateUser
// @description: Validate user credentials and permissions
func ValidateUser(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("👤 Validating user with User Service...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	orderRequest, ok := dataMap["order_request"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("order_request not found")
	}

	userID, ok := orderRequest["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found")
	}

	// Simulate API call delay
	time.Sleep(100 * time.Millisecond)

	// Simulate validation logic
	if userID == "" {
		return nil, fmt.Errorf("invalid user ID")
	}

	fmt.Printf("   ✅ User %s validated successfully\n", userID)

	// Add validation result to data
	dataMap["user_validated"] = true
	dataMap["user_permissions"] = []string{"order", "payment", "shipping"}

	return dataMap, nil
}

// @workflow:task
// @name: checkInventory
// @description: Check product availability in inventory service
func CheckInventory(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("📦 Checking inventory with Inventory Service...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	orderRequest, ok := dataMap["order_request"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("order_request not found")
	}

	productID, ok := orderRequest["product_id"].(string)
	if !ok {
		return nil, fmt.Errorf("product_id not found")
	}

	quantity, ok := orderRequest["quantity"].(int)
	if !ok {
		// Try to convert from float64 (JSON number)
		if q, ok := orderRequest["quantity"].(float64); ok {
			quantity = int(q)
		} else {
			return nil, fmt.Errorf("quantity not found or invalid")
		}
	}

	// Simulate API call delay
	time.Sleep(150 * time.Millisecond)

	// Simulate inventory check
	availableStock := rand.Intn(10) + 5 // 5-14 items available

	if quantity > availableStock {
		return nil, fmt.Errorf("insufficient inventory: requested %d, available %d", quantity, availableStock)
	}

	fmt.Printf("   ✅ Product %s available - requested: %d, available: %d\n", productID, quantity, availableStock)

	// Add inventory info to data
	dataMap["inventory_check"] = map[string]interface{}{
		"product_id":      productID,
		"requested":       quantity,
		"available":       availableStock,
		"check_timestamp": time.Now().Unix(),
	}

	return dataMap, nil
}

// @workflow:task
// @name: reserveInventory
// @description: Reserve items in inventory for the order
func ReserveInventory(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🔒 Reserving inventory with Inventory Service...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	inventoryCheck, ok := dataMap["inventory_check"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("inventory_check not found")
	}

	productID := inventoryCheck["product_id"].(string)
	quantity := inventoryCheck["requested"].(int)

	// Simulate API call delay
	time.Sleep(200 * time.Millisecond)

	// Generate reservation ID
	reservationID := fmt.Sprintf("RES-%d-%s", time.Now().Unix(), productID)

	fmt.Printf("   ✅ Reserved %d units of %s - Reservation ID: %s\n", quantity, productID, reservationID)

	// Add reservation info to data
	dataMap["inventory_reservation"] = map[string]interface{}{
		"reservation_id": reservationID,
		"product_id":     productID,
		"quantity":       quantity,
		"reserved_at":    time.Now().Unix(),
		"expires_at":     time.Now().Add(15 * time.Minute).Unix(),
	}

	return dataMap, nil
}

// @workflow:task
// @name: processPayment
// @description: Process payment through payment service
func ProcessPayment(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("💳 Processing payment with Payment Service...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	orderRequest, ok := dataMap["order_request"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("order_request not found")
	}

	amount, ok := orderRequest["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("amount not found or invalid")
	}

	userID := orderRequest["user_id"].(string)

	// Simulate payment processing delay
	time.Sleep(300 * time.Millisecond)

	// Simulate payment processing (90% success rate)
	if rand.Float32() > 0.9 {
		return nil, fmt.Errorf("payment failed: insufficient funds")
	}

	// Generate transaction ID
	transactionID := fmt.Sprintf("TXN-%d-%s", time.Now().Unix(), userID)

	fmt.Printf("   ✅ Payment processed - Amount: $%.2f, Transaction ID: %s\n", amount, transactionID)

	// Add payment info to data
	dataMap["payment"] = map[string]interface{}{
		"transaction_id": transactionID,
		"amount":         amount,
		"currency":       "USD",
		"status":         "completed",
		"processed_at":   time.Now().Unix(),
		"payment_method": "credit_card",
	}

	return dataMap, nil
}

// @workflow:task
// @name: createShipment
// @description: Create shipment in logistics service
func CreateShipment(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🚚 Creating shipment with Logistics Service...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	orderRequest, ok := dataMap["order_request"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("order_request not found")
	}

	shippingAddress, ok := orderRequest["shipping_address"].(map[string]string)
	if !ok {
		return nil, fmt.Errorf("shipping_address not found")
	}

	orderID := orderRequest["order_id"].(string)

	// Simulate API call delay
	time.Sleep(250 * time.Millisecond)

	// Generate tracking number
	trackingNumber := fmt.Sprintf("TRK-%d-%s", time.Now().Unix(), orderID)
	estimatedDelivery := time.Now().Add(3 * 24 * time.Hour) // 3 days from now

	fmt.Printf("   ✅ Shipment created - Tracking: %s, Estimated delivery: %s\n",
		trackingNumber, estimatedDelivery.Format("2006-01-02"))

	// Add shipment info to data
	dataMap["shipment"] = map[string]interface{}{
		"tracking_number":    trackingNumber,
		"carrier":            "FastShip Express",
		"shipping_address":   shippingAddress,
		"estimated_delivery": estimatedDelivery.Unix(),
		"created_at":         time.Now().Unix(),
		"status":             "created",
	}

	return dataMap, nil
}

// @workflow:task
// @name: sendNotification
// @description: Send confirmation notification to customer
func SendNotification(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("📧 Sending notification with Notification Service...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	orderRequest, ok := dataMap["order_request"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("order_request not found")
	}

	payment, ok := dataMap["payment"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("payment not found")
	}

	shipment, ok := dataMap["shipment"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("shipment not found")
	}

	orderID := orderRequest["order_id"].(string)
	userID := orderRequest["user_id"].(string)
	transactionID := payment["transaction_id"].(string)
	trackingNumber := shipment["tracking_number"].(string)

	// Simulate notification sending delay
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("   ✅ Notification sent to user %s\n", userID)
	fmt.Printf("      📧 Email: Order %s confirmed - Transaction: %s\n", orderID, transactionID)
	fmt.Printf("      📱 SMS: Your order is being shipped - Tracking: %s\n", trackingNumber)

	// Add notification info to data
	dataMap["notification"] = map[string]interface{}{
		"email_sent": true,
		"sms_sent":   true,
		"push_sent":  true,
		"sent_at":    time.Now().Unix(),
		"channels":   []string{"email", "sms", "push"},
	}

	// Add final order summary
	dataMap["order_summary"] = map[string]interface{}{
		"order_id":        orderID,
		"user_id":         userID,
		"status":          "completed",
		"transaction_id":  transactionID,
		"tracking_number": trackingNumber,
		"completed_at":    time.Now().Unix(),
	}

	return dataMap, nil
}

// Helper functions

type ProductInventory struct {
	Name        string
	Price       float64
	Stock       int
	Available   int
	Reserved    int
	WarehouseID string
}

func getProductInventory(productID string) ProductInventory {
	inventoryMap := map[string]ProductInventory{
		"PROD_001": {
			Name:        "Premium Wireless Headphones",
			Price:       299.99,
			Stock:       50,
			WarehouseID: "WH_001",
		},
		"PROD_002": {
			Name:        "Smart Fitness Watch",
			Price:       199.99,
			Stock:       25,
			WarehouseID: "WH_002",
		},
		"PROD_003": {
			Name:        "Bluetooth Speaker",
			Price:       89.99,
			Stock:       100,
			WarehouseID: "WH_001",
		},
	}

	if inventory, exists := inventoryMap[productID]; exists {
		return inventory
	}

	// Default product
	return ProductInventory{
		Name:        "Generic Product",
		Price:       49.99,
		Stock:       10,
		WarehouseID: "WH_003",
	}
}

func getUserTier(status string) string {
	switch status {
	case "premium":
		return "platinum"
	case "verified":
		return "gold"
	default:
		return "silver"
	}
}

func getDiscountByTier(tier string) float64 {
	switch tier {
	case "platinum":
		return 0.15 // 15% discount
	case "gold":
		return 0.10 // 10% discount
	case "silver":
		return 0.05 // 5% discount
	default:
		return 0.0
	}
}
