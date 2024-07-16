package uses_cases

import "fmt"

// Task2
// @type: task
// @description Execute Task 2.
//
// @input data (interface{}): Input data.
//
// @output string: Returns result from task 2.
// @output error: Returns an error if occurs.
func Task2(data interface{}) (string, error) {
	fmt.Println("Executing Task 2")
	return "result from task 2", nil
}
