package uses_cases

import "fmt"

// Task1
// @type: task
// @description Execute Task 1.
//
// @input data (interface{}): Input data.
//
// @output string: Returns result from task 1.
// @output error: Returns an error if occurs.
func Task1(data interface{}) (string, error) {
	fmt.Println("Executing Task 1")
	return "result from task 1", nil
}
