package uses_cases

import "fmt"

// Task2
// @type: task
// @description Execute Task 2.
//
// @input data (interface{}): Input data.
//
// @output interface{}: Returns nil.
// @output error: Returns an error if occurs.
func Task2(data interface{}) (interface{}, error) {
	fmt.Println("Executing Task 2 with data:", data)
	return nil, nil
}

// BeforeTask2
// @type: pre-hook
// @before Task2
// @description Execute before Task 2.
func BeforeTask2(data interface{}) (interface{}, error) {
	fmt.Println("Starting task 2")
	fmt.Println("Executing Before Task 2", data)
	return "data", nil
}
