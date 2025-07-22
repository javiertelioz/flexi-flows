package tasks

import (
	"context"
	"fmt"
	"strings"
)

// @workflow:task
// @name: validateData
// @description: Validates user input data
func ValidateData(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format, expected map")
	}

	user, ok := dataMap["user"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("user data not found or invalid format")
	}

	// Validate required fields
	name, nameOk := user["name"].(string)
	email, emailOk := user["email"].(string)
	age, ageOk := user["age"].(float64)

	if !nameOk || name == "" {
		return nil, fmt.Errorf("name is required and must be a string")
	}

	if !emailOk || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("valid email is required")
	}

	if !ageOk || age < 0 || age > 150 {
		return nil, fmt.Errorf("age must be between 0 and 150")
	}

	// Return validated data
	result := map[string]interface{}{
		"validated": true,
		"user":      user,
		"validation_passed": map[string]bool{
			"name":  true,
			"email": true,
			"age":   true,
		},
	}

	fmt.Printf("✅ Data validated successfully for user: %s\n", name)
	return result, nil
}

// @workflow:task
// @name: transformData
// @description: Transforms user data to required format
func TransformData(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	user, ok := dataMap["user"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("user data not found")
	}

	name := user["name"].(string)
	email := user["email"].(string)
	age := user["age"].(float64)

	// Transform data
	transformed := map[string]interface{}{
		"fullName": strings.ToUpper(name),
		"email":    email,
		"category": func() string {
			if age >= 18 {
				return "adult"
			}
			return "minor"
		}(),
		"ageGroup": func() string {
			switch {
			case age < 18:
				return "youth"
			case age < 65:
				return "adult"
			default:
				return "senior"
			}
		}(),
	}

	result := map[string]interface{}{
		"validated":   dataMap["validated"],
		"original":    user,
		"transformed": transformed,
	}

	fmt.Printf("🔄 Data transformed for user: %s -> %s\n", name, transformed["fullName"])
	return result, nil
}

// @workflow:task
// @name: saveData
// @description: Saves processed data (simulated)
func SaveData(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	transformed, ok := dataMap["transformed"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("transformed data not found")
	}

	// Simulate saving to database
	userID := fmt.Sprintf("user_%d", int(len(transformed["fullName"].(string))))

	result := map[string]interface{}{
		"validated":   dataMap["validated"],
		"original":    dataMap["original"],
		"transformed": transformed,
		"saved":       true,
		"userID":      userID,
		"savedAt":     "2025-07-21T10:30:00Z",
	}

	fmt.Printf("💾 Data saved successfully with ID: %s\n", userID)
	return result, nil
}

// @workflow:task
// @name: notifyUser
// @description: Sends notification to user (simulated)
func NotifyUser(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	transformed, ok := dataMap["transformed"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("transformed data not found")
	}

	email := transformed["email"].(string)
	fullName := transformed["fullName"].(string)

	// Simulate sending notification
	notification := map[string]interface{}{
		"to":      email,
		"subject": fmt.Sprintf("Welcome %s!", fullName),
		"message": "Your account has been successfully created and processed.",
		"sent":    true,
		"sentAt":  "2025-07-21T10:31:00Z",
	}

	result := map[string]interface{}{
		"validated":    dataMap["validated"],
		"original":     dataMap["original"],
		"transformed":  transformed,
		"saved":        dataMap["saved"],
		"userID":       dataMap["userID"],
		"savedAt":      dataMap["savedAt"],
		"notified":     true,
		"notification": notification,
	}

	fmt.Printf("📧 Notification sent to: %s\n", email)
	return result, nil
}
