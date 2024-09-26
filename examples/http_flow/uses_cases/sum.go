package use_cases

import (
	"fmt"
	"log"
)

// SumFunc
// @type: task
// @description calculates the sum of prime status and square.
//
// @input data ([]interface{}): The results of prime check and square calculation.
//
// @output int: Returns the sum of the results.
// @output error: Returns an error if occurs.
func SumFunc(data any) (any, error) {
	results, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("SumFunc expects data of type map[string]interface{}, got %T", data)
	}

	isPrime, ok := results["isPrime"].(bool)
	if !ok {
		return nil, fmt.Errorf("isPrime result not found or invalid type")
	}
	square, ok := results["square"].(int)
	if !ok {
		return nil, fmt.Errorf("square result not found or invalid type")
	}

	sum := square
	if isPrime {
		sum += 1
	}
	log.Printf("Sum of results: %d (Prime: %v, Square: %d)", sum, isPrime, square)
	return sum, nil
}

// BeforeSum
// @type: pre-hook
// @before SumFunc
// @description execute before SumFunc.
//
// @input data ([]interface{}): The results of prime check and square calculation.
//
// @output []interface{}: Returns data to be summed.
// @output error: Returns an error if occurs.
func BeforeSum(data any) (any, error) {
	log.Printf("Before sum: received data %v", data)
	return data, nil
}

// AfterSum
// @type: post-hook
// @after SumFunc
// @description execute after SumFunc.
//
// @input result (int): The result of the SumFunc.
//
// @output AfterSumResult: Returns the result after post-processing.
func AfterSum(result any) (any, error) {
	sum, ok := result.(int)
	if !ok {
		return nil, fmt.Errorf("AfterSum expects result of type int, got %T", result)
	}
	sum = sum - 19
	log.Printf("After sum result is: %d", sum)
	return sum, nil
}
