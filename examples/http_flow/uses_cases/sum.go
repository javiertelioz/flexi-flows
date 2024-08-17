package use_cases

import "fmt"

// SumFunc
// @type: task
// @description calculates the sum of prime status and square.
//
// @input data ([]interface{}): The results of prime check and square calculation.
//
// @output int: Returns the sum of the results.
// @output error: Returns an error if occurs.
func SumFunc(data []interface{}) (int, error) {
	isPrime := data[0].(bool)
	square := data[1].(int)
	sum := square
	if isPrime {
		sum += 1
	}
	fmt.Printf("Sum of results: %d (Prime: %v, Square: %d)\n", sum, isPrime, square)
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
func BeforeSum(data []interface{}) ([]interface{}, error) {
	fmt.Println("Before sum: received data", data)
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
// @output error: Returns an error if occurs.
func AfterSum(result int) (int, error) {
	result = result - 19
	fmt.Printf("After sum result is: %d\n", result)
	return result, nil
}
