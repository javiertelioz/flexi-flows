package tasks

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// @workflow:task
// @name: extractData
// @description: Extract data from various sources (API, mock, database)
func ExtractData(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("📥 Extracting data from source...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	source, _ := dataMap["source"].(string)
	batchSize, _ := dataMap["batch_size"].(int)
	if batchSize == 0 {
		batchSize = 5
	}

	// Simulate data extraction based on source
	var extractedData []map[string]interface{}

	switch source {
	case "api":
		// Simulate API data extraction
		extractedData = generateAPIData(batchSize)
		fmt.Printf("   ✅ Extracted %d records from API\n", len(extractedData))
	case "mock":
		// Generate mock data
		extractedData = generateMockData(batchSize)
		fmt.Printf("   ✅ Generated %d mock records\n", len(extractedData))
	default:
		// Default to mock data
		extractedData = generateMockData(batchSize)
		fmt.Printf("   ✅ Generated %d default records\n", len(extractedData))
	}

	result := map[string]interface{}{
		"source":          source,
		"records":         extractedData,
		"extraction_time": time.Now().Unix(),
		"total_records":   len(extractedData),
		"batch_size":      batchSize,
	}

	// Simulate extraction delay
	time.Sleep(100 * time.Millisecond)

	return result, nil
}

// @workflow:task
// @name: validateData
// @description: Validate extracted data for completeness and correctness
func ValidateData(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("✅ Validating data integrity...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	records, ok := dataMap["records"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid records format")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	validRecords := []map[string]interface{}{}
	invalidRecords := []map[string]interface{}{}
	validationStats := map[string]int{
		"total":         len(records),
		"valid":         0,
		"invalid_email": 0,
		"missing_name":  0,
		"invalid_age":   0,
	}

	for _, record := range records {
		isValid := true
		validationErrors := []string{}

		// Validate email
		if email, exists := record["email"].(string); !exists || !emailRegex.MatchString(email) {
			isValid = false
			validationErrors = append(validationErrors, "invalid_email")
			validationStats["invalid_email"]++
		}

		// Validate name
		if name, exists := record["name"].(string); !exists || strings.TrimSpace(name) == "" {
			isValid = false
			validationErrors = append(validationErrors, "missing_name")
			validationStats["missing_name"]++
		}

		// Validate age
		if age, exists := record["age"]; exists {
			if ageFloat, ok := age.(float64); !ok || ageFloat < 0 || ageFloat > 120 {
				isValid = false
				validationErrors = append(validationErrors, "invalid_age")
				validationStats["invalid_age"]++
			}
		}

		if isValid {
			validRecords = append(validRecords, record)
			validationStats["valid"]++
		} else {
			record["validation_errors"] = validationErrors
			invalidRecords = append(invalidRecords, record)
		}
	}

	result := map[string]interface{}{
		"valid_records":    validRecords,
		"invalid_records":  invalidRecords,
		"validation_stats": validationStats,
		"validation_time":  time.Now().Unix(),
		"original_data":    dataMap,
	}

	fmt.Printf("   ✅ Validation completed: %d valid, %d invalid records\n",
		validationStats["valid"], len(invalidRecords))

	time.Sleep(50 * time.Millisecond)
	return result, nil
}

// @workflow:task
// @name: transformData
// @description: Transform and normalize data fields
func TransformData(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🔄 Transforming and normalizing data...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	validRecords, ok := dataMap["valid_records"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid valid_records format")
	}

	transformedRecords := []map[string]interface{}{}

	for _, record := range validRecords {
		transformedRecord := make(map[string]interface{})

		// Copy original fields
		for k, v := range record {
			transformedRecord[k] = v
		}

		// Normalize name (Title Case)
		if name, exists := record["name"].(string); exists {
			transformedRecord["name_normalized"] = strings.Title(strings.ToLower(strings.TrimSpace(name)))
		}

		// Extract domain from email
		if email, exists := record["email"].(string); exists {
			if parts := strings.Split(email, "@"); len(parts) == 2 {
				transformedRecord["email_domain"] = parts[1]
			}
		}

		// Calculate age group
		if age, exists := record["age"].(float64); exists {
			transformedRecord["age_group"] = calculateAgeGroup(int(age))
		}

		// Add transformation timestamp
		transformedRecord["transformed_at"] = time.Now().Unix()

		transformedRecords = append(transformedRecords, transformedRecord)
	}

	result := map[string]interface{}{
		"transformed_records": transformedRecords,
		"transformation_time": time.Now().Unix(),
		"records_processed":   len(transformedRecords),
		"original_data":       dataMap,
	}

	fmt.Printf("   ✅ Transformed %d records\n", len(transformedRecords))

	time.Sleep(75 * time.Millisecond)
	return result, nil
}

// @workflow:task
// @name: enrichData
// @description: Enrich data with additional information
func EnrichData(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🌟 Enriching data with additional information...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	transformedRecords, ok := dataMap["transformed_records"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid transformed_records format")
	}

	enrichedRecords := []map[string]interface{}{}

	for _, record := range transformedRecords {
		enrichedRecord := make(map[string]interface{})

		// Copy all existing fields
		for k, v := range record {
			enrichedRecord[k] = v
		}

		// Add user score based on various factors
		score := calculateUserScore(record)
		enrichedRecord["user_score"] = score

		// Add location info based on email domain (simulated)
		if domain, exists := record["email_domain"].(string); exists {
			enrichedRecord["estimated_location"] = getLocationByDomain(domain)
		}

		// Add risk assessment
		enrichedRecord["risk_level"] = assessRiskLevel(record)

		// Add enrichment metadata
		enrichedRecord["enriched_at"] = time.Now().Unix()
		enrichedRecord["enrichment_version"] = "1.0"

		enrichedRecords = append(enrichedRecords, enrichedRecord)
	}

	result := map[string]interface{}{
		"enriched_records": enrichedRecords,
		"enrichment_time":  time.Now().Unix(),
		"records_enriched": len(enrichedRecords),
		"original_data":    dataMap,
	}

	fmt.Printf("   ✅ Enriched %d records with additional data\n", len(enrichedRecords))

	time.Sleep(100 * time.Millisecond)
	return result, nil
}

// @workflow:task
// @name: filterData
// @description: Filter data based on business rules
func FilterData(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("🔍 Filtering data based on business rules...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	enrichedRecords, ok := dataMap["enriched_records"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid enriched_records format")
	}

	filteredRecords := []map[string]interface{}{}
	excludedRecords := []map[string]interface{}{}

	for _, record := range enrichedRecords {
		include := true
		exclusionReasons := []string{}

		// Filter by user score (minimum score of 3.0)
		if score, exists := record["user_score"].(float64); exists && score < 3.0 {
			include = false
			exclusionReasons = append(exclusionReasons, "low_user_score")
		}

		// Filter by risk level (exclude high risk)
		if riskLevel, exists := record["risk_level"].(string); exists && riskLevel == "high" {
			include = false
			exclusionReasons = append(exclusionReasons, "high_risk")
		}

		// Filter by age (must be 18+)
		if age, exists := record["age"].(float64); exists && age < 18 {
			include = false
			exclusionReasons = append(exclusionReasons, "under_age")
		}

		if include {
			filteredRecords = append(filteredRecords, record)
		} else {
			record["exclusion_reasons"] = exclusionReasons
			excludedRecords = append(excludedRecords, record)
		}
	}

	result := map[string]interface{}{
		"filtered_records": filteredRecords,
		"excluded_records": excludedRecords,
		"filter_time":      time.Now().Unix(),
		"records_passed":   len(filteredRecords),
		"records_excluded": len(excludedRecords),
		"original_data":    dataMap,
	}

	fmt.Printf("   ✅ Filtering completed: %d records passed, %d excluded\n",
		len(filteredRecords), len(excludedRecords))

	time.Sleep(50 * time.Millisecond)
	return result, nil
}

// @workflow:task
// @name: aggregateData
// @description: Aggregate and summarize processed data
func AggregateData(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("📊 Aggregating data and generating summaries...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	filteredRecords, ok := dataMap["filtered_records"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid filtered_records format")
	}

	// Initialize aggregation counters
	aggregations := map[string]interface{}{
		"total_processed":    len(filteredRecords),
		"age_groups":         make(map[string]int),
		"email_domains":      make(map[string]int),
		"locations":          make(map[string]int),
		"risk_levels":        make(map[string]int),
		"average_score":      0.0,
		"score_distribution": make(map[string]int),
	}

	totalScore := 0.0
	ageGroups := aggregations["age_groups"].(map[string]int)
	emailDomains := aggregations["email_domains"].(map[string]int)
	locations := aggregations["locations"].(map[string]int)
	riskLevels := aggregations["risk_levels"].(map[string]int)
	scoreDistribution := aggregations["score_distribution"].(map[string]int)

	// Process each record for aggregation
	for _, record := range filteredRecords {
		// Age group aggregation
		if ageGroup, exists := record["age_group"].(string); exists {
			ageGroups[ageGroup]++
		}

		// Email domain aggregation
		if domain, exists := record["email_domain"].(string); exists {
			emailDomains[domain]++
		}

		// Location aggregation
		if location, exists := record["estimated_location"].(string); exists {
			locations[location]++
		}

		// Risk level aggregation
		if riskLevel, exists := record["risk_level"].(string); exists {
			riskLevels[riskLevel]++
		}

		// Score aggregation
		if score, exists := record["user_score"].(float64); exists {
			totalScore += score
			scoreRange := getScoreRange(score)
			scoreDistribution[scoreRange]++
		}
	}

	// Calculate average score
	if len(filteredRecords) > 0 {
		aggregations["average_score"] = totalScore / float64(len(filteredRecords))
	}

	result := map[string]interface{}{
		"aggregations":      aggregations,
		"aggregation_time":  time.Now().Unix(),
		"processed_records": filteredRecords,
		"original_data":     dataMap,
	}

	fmt.Printf("   ✅ Aggregated %d records with statistical summaries\n", len(filteredRecords))

	time.Sleep(75 * time.Millisecond)
	return result, nil
}

// @workflow:task
// @name: saveData
// @description: Save processed data to storage
func SaveData(ctx context.Context, data interface{}) (interface{}, error) {
	fmt.Println("💾 Saving processed data to storage...")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	processedRecords, ok := dataMap["processed_records"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid processed_records format")
	}

	aggregations, _ := dataMap["aggregations"].(map[string]interface{})

	// Simulate saving to different storage systems
	saveResults := map[string]interface{}{
		"database": map[string]interface{}{
			"status":         "success",
			"records_saved":  len(processedRecords),
			"table":          "processed_users",
			"transaction_id": fmt.Sprintf("TXN_%d", time.Now().Unix()),
		},
		"cache": map[string]interface{}{
			"status":      "success",
			"cache_key":   fmt.Sprintf("aggregations_%d", time.Now().Unix()),
			"ttl_seconds": 3600,
		},
		"export": map[string]interface{}{
			"status":   "success",
			"filename": fmt.Sprintf("processed_data_%d.json", time.Now().Unix()),
			"format":   "json",
			"size_mb":  float64(len(processedRecords)) * 0.1, // Simulated size
		},
	}

	result := map[string]interface{}{
		"save_results":    saveResults,
		"saved_at":        time.Now().Unix(),
		"records_count":   len(processedRecords),
		"aggregations":    aggregations,
		"pipeline_status": "completed",
		"processing_time": calculateProcessingTime(dataMap),
	}

	fmt.Printf("   ✅ Successfully saved %d records to storage systems\n", len(processedRecords))

	// Simulate save delay
	time.Sleep(125 * time.Millisecond)

	return result, nil
}

// Helper functions

func generateMockData(count int) []map[string]interface{} {
	names := []string{"Alice Johnson", "Bob Smith", "Carol Davis", "David Wilson", "Eva Brown", "Frank Miller", "Grace Lee", "Henry Taylor", "Ivy Chen", "Jack Anderson"}
	domains := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "company.com", "university.edu"}

	records := make([]map[string]interface{}, count)

	for i := 0; i < count; i++ {
		name := names[rand.Intn(len(names))]
		domain := domains[rand.Intn(len(domains))]
		email := fmt.Sprintf("%s@%s",
			strings.ToLower(strings.ReplaceAll(name, " ", ".")),
			domain)

		records[i] = map[string]interface{}{
			"id":    i + 1,
			"name":  name,
			"email": email,
			"age":   float64(rand.Intn(60) + 18), // Age between 18-78
		}
	}

	return records
}

func generateAPIData(count int) []map[string]interface{} {
	// Simulate API data with slightly different structure
	return generateMockData(count)
}

func calculateAgeGroup(age int) string {
	switch {
	case age < 25:
		return "young_adult"
	case age < 35:
		return "adult"
	case age < 50:
		return "middle_aged"
	default:
		return "senior"
	}
}

func calculateUserScore(record map[string]interface{}) float64 {
	score := 5.0 // Base score

	// Adjust based on age
	if age, exists := record["age"].(float64); exists {
		if age >= 25 && age <= 45 {
			score += 1.0
		} else if age > 65 {
			score -= 0.5
		}
	}

	// Adjust based on email domain
	if domain, exists := record["email_domain"].(string); exists {
		if strings.Contains(domain, "company.com") || strings.Contains(domain, "university.edu") {
			score += 1.5
		}
	}

	// Add some randomness
	score += (rand.Float64() - 0.5) * 2.0 // Random adjustment between -1.0 and +1.0

	// Ensure score is between 1.0 and 10.0
	if score < 1.0 {
		score = 1.0
	} else if score > 10.0 {
		score = 10.0
	}

	return score
}

func getLocationByDomain(domain string) string {
	locationMap := map[string]string{
		"gmail.com":      "US",
		"yahoo.com":      "US",
		"hotmail.com":    "US",
		"outlook.com":    "US",
		"company.com":    "Corporate",
		"university.edu": "Academic",
	}

	if location, exists := locationMap[domain]; exists {
		return location
	}
	return "Unknown"
}

func assessRiskLevel(record map[string]interface{}) string {
	risk := 0

	// Age factor
	if age, exists := record["age"].(float64); exists {
		if age < 21 || age > 70 {
			risk++
		}
	}

	// Score factor
	if score, exists := record["user_score"].(float64); exists {
		if score < 4.0 {
			risk++
		}
	}

	switch risk {
	case 0:
		return "low"
	case 1:
		return "medium"
	default:
		return "high"
	}
}

func getScoreRange(score float64) string {
	switch {
	case score < 3.0:
		return "low (1-3)"
	case score < 6.0:
		return "medium (3-6)"
	case score < 8.0:
		return "high (6-8)"
	default:
		return "excellent (8-10)"
	}
}

func calculateProcessingTime(dataMap map[string]interface{}) float64 {
	// Calculate total processing time by looking at timestamps
	if originalData, exists := dataMap["original_data"].(map[string]interface{}); exists {
		if extractionTime, exists := originalData["extraction_time"].(int64); exists {
			return float64(time.Now().Unix() - extractionTime)
		}
	}
	return 0.0
}
