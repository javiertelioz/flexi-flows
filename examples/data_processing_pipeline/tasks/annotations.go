package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// User represents a user from the API
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
	Company  struct {
		Name string `json:"name"`
	} `json:"company"`
}

// @workflow:task
// @name: fetchUsers
// @description: Fetches users from external API or mock data
func FetchUsers(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	source, _ := dataMap["source"].(string)
	if source == "" {
		source = "api"
	}

	var users []User
	var err error

	switch source {
	case "api":
		users, err = fetchFromAPI(ctx)
	case "mock":
		users = getMockUsers()
	default:
		return nil, fmt.Errorf("unknown source: %s", source)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	result := map[string]interface{}{
		"users":       users,
		"total_count": len(users),
		"source":      source,
		"fetched_at":  time.Now().Format(time.RFC3339),
	}

	fmt.Printf("📥 Fetched %d users from %s\n", len(users), source)
	return result, nil
}

// @workflow:task
// @name: validateUsers
// @description: Validates each user in the collection
func ValidateUsers(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	usersData, ok := dataMap["users"].([]interface{})
	if !ok {
		// Try to convert from []User
		if usersSlice, ok := dataMap["users"].([]User); ok {
			usersData = make([]interface{}, len(usersSlice))
			for i, user := range usersSlice {
				usersData[i] = user
			}
		} else {
			return nil, fmt.Errorf("users data not found or invalid format")
		}
	}

	validUsers := []interface{}{}
	invalidUsers := []interface{}{}
	validationErrors := []string{}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	for i, userInterface := range usersData {
		// Convert interface{} to User
		userBytes, err := json.Marshal(userInterface)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("User %d: failed to marshal", i))
			continue
		}

		var user User
		err = json.Unmarshal(userBytes, &user)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("User %d: failed to unmarshal", i))
			continue
		}

		// Validate user
		isValid := true
		errors := []string{}

		if user.Name == "" {
			isValid = false
			errors = append(errors, "name is required")
		}

		if user.Email == "" || !emailRegex.MatchString(user.Email) {
			isValid = false
			errors = append(errors, "valid email is required")
		}

		if user.Username == "" {
			isValid = false
			errors = append(errors, "username is required")
		}

		if isValid {
			validUsers = append(validUsers, user)
		} else {
			invalidUser := map[string]interface{}{
				"user":   user,
				"errors": errors,
			}
			invalidUsers = append(invalidUsers, invalidUser)
			validationErrors = append(validationErrors, fmt.Sprintf("User %s: %s", user.Name, strings.Join(errors, ", ")))
		}
	}

	result := map[string]interface{}{
		"valid_users":       validUsers,
		"invalid_users":     invalidUsers,
		"validation_errors": validationErrors,
		"total_count":       dataMap["total_count"],
		"valid_count":       len(validUsers),
		"invalid_count":     len(invalidUsers),
		"source":            dataMap["source"],
		"fetched_at":        dataMap["fetched_at"],
		"validated_at":      time.Now().Format(time.RFC3339),
	}

	fmt.Printf("✅ Validated users: %d valid, %d invalid\n", len(validUsers), len(invalidUsers))
	return result, nil
}

// @workflow:task
// @name: transformUsers
// @description: Transforms valid users to required format
func TransformUsers(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	validUsersData, ok := dataMap["valid_users"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("valid users data not found")
	}

	transformedUsers := []interface{}{}
	domains := make(map[string]int)

	for _, userInterface := range validUsersData {
		// Convert interface{} to User
		userBytes, err := json.Marshal(userInterface)
		if err != nil {
			continue
		}

		var user User
		err = json.Unmarshal(userBytes, &user)
		if err != nil {
			continue
		}

		// Extract domain from email
		emailParts := strings.Split(user.Email, "@")
		domain := ""
		if len(emailParts) == 2 {
			domain = emailParts[1]
			domains[domain]++
		}

		// Transform user
		transformed := map[string]interface{}{
			"id":            user.ID,
			"full_name":     strings.Title(strings.ToLower(user.Name)),
			"username":      strings.ToLower(user.Username),
			"email":         strings.ToLower(user.Email),
			"domain":        domain,
			"phone":         user.Phone,
			"website":       user.Website,
			"company_name":  user.Company.Name,
			"profile_score": calculateProfileScore(user),
		}

		transformedUsers = append(transformedUsers, transformed)
	}

	// Convert domains map to slice for easier JSON handling
	domainList := []string{}
	for domain := range domains {
		domainList = append(domainList, domain)
	}

	result := map[string]interface{}{
		"transformed_users": transformedUsers,
		"domains_found":     domainList,
		"domain_counts":     domains,
		"total_count":       dataMap["total_count"],
		"valid_count":       dataMap["valid_count"],
		"invalid_count":     dataMap["invalid_count"],
		"invalid_users":     dataMap["invalid_users"],
		"validation_errors": dataMap["validation_errors"],
		"source":            dataMap["source"],
		"fetched_at":        dataMap["fetched_at"],
		"validated_at":      dataMap["validated_at"],
		"transformed_at":    time.Now().Format(time.RFC3339),
	}

	fmt.Printf("🔄 Transformed %d users, found %d unique domains\n", len(transformedUsers), len(domainList))
	return result, nil
}

// @workflow:task
// @name: saveUsers
// @description: Saves processed users (simulated database save)
func SaveUsers(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	transformedUsers, ok := dataMap["transformed_users"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("transformed users data not found")
	}

	// Simulate batch save operation
	batchID := fmt.Sprintf("batch_%d", time.Now().Unix())
	saved := 0

	for range transformedUsers {
		// Simulate save with small delay
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		time.Sleep(10 * time.Millisecond) // Simulate DB write time
		saved++
	}

	result := map[string]interface{}{
		"batch_id":          batchID,
		"users_saved":       saved,
		"transformed_users": transformedUsers,
		"domains_found":     dataMap["domains_found"],
		"domain_counts":     dataMap["domain_counts"],
		"total_count":       dataMap["total_count"],
		"valid_count":       dataMap["valid_count"],
		"invalid_count":     dataMap["invalid_count"],
		"invalid_users":     dataMap["invalid_users"],
		"validation_errors": dataMap["validation_errors"],
		"source":            dataMap["source"],
		"fetched_at":        dataMap["fetched_at"],
		"validated_at":      dataMap["validated_at"],
		"transformed_at":    dataMap["transformed_at"],
		"saved_at":          time.Now().Format(time.RFC3339),
	}

	fmt.Printf("💾 Saved %d users with batch ID: %s\n", saved, batchID)
	return result, nil
}

// @workflow:task
// @name: generateReport
// @description: Generates processing report
func GenerateReport(ctx context.Context, data interface{}) (interface{}, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	report := map[string]interface{}{
		"processing_summary": map[string]interface{}{
			"total_users":   dataMap["total_count"],
			"valid_users":   dataMap["valid_count"],
			"invalid_users": dataMap["invalid_count"],
			"saved_users":   dataMap["users_saved"],
			"domains_found": dataMap["domains_found"],
			"batch_id":      dataMap["batch_id"],
		},
		"timestamps": map[string]interface{}{
			"fetched_at":     dataMap["fetched_at"],
			"validated_at":   dataMap["validated_at"],
			"transformed_at": dataMap["transformed_at"],
			"saved_at":       dataMap["saved_at"],
			"report_at":      time.Now().Format(time.RFC3339),
		},
		"validation_errors": dataMap["validation_errors"],
		"source":            dataMap["source"],
		"report_generated":  true,
	}

	// Add all previous data for complete result
	for key, value := range dataMap {
		if key != "validation_errors" && key != "source" {
			report[key] = value
		}
	}

	fmt.Printf("📊 Generated processing report for batch: %s\n", dataMap["batch_id"])
	return report, nil
}

// Helper functions
func fetchFromAPI(ctx context.Context) ([]User, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://jsonplaceholder.typicode.com/users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var users []User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}

func getMockUsers() []User {
	return []User{
		{ID: 1, Name: "John Doe", Username: "johndoe", Email: "john@example.com", Phone: "123-456-7890", Website: "john.com", Company: struct {
			Name string `json:"name"`
		}{Name: "Example Corp"}},
		{ID: 2, Name: "Jane Smith", Username: "janesmith", Email: "jane@gmail.com", Phone: "098-765-4321", Website: "jane.org", Company: struct {
			Name string `json:"name"`
		}{Name: "Tech Solutions"}},
		{ID: 3, Name: "", Username: "invalid", Email: "invalid-email", Phone: "", Website: "", Company: struct {
			Name string `json:"name"`
		}{Name: ""}}, // Invalid user
		{ID: 4, Name: "Bob Wilson", Username: "bobwilson", Email: "bob@yahoo.com", Phone: "555-0123", Website: "bob.net", Company: struct {
			Name string `json:"name"`
		}{Name: "Wilson Ltd"}},
		{ID: 5, Name: "Alice Brown", Username: "", Email: "alice@domain.com", Phone: "444-0987", Website: "", Company: struct {
			Name string `json:"name"`
		}{Name: "Brown Industries"}}, // Invalid user
	}
}

func calculateProfileScore(user User) int {
	score := 0

	if user.Name != "" {
		score += 20
	}
	if user.Email != "" {
		score += 25
	}
	if user.Phone != "" {
		score += 15
	}
	if user.Website != "" {
		score += 20
	}
	if user.Company.Name != "" {
		score += 20
	}

	return score
}
