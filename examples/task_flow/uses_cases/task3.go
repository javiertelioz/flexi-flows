package uses_cases

import "fmt"

// Task3
// @type: task
// @description Execute Task 3.
//
// @input data (interface{}): Input data.
//
// @output string: Returns result from task 3.
// @output error: Returns an error if occurs.
func Task3(data interface{}) (string, error) {
	fmt.Println("Executing Task 3")
	return "result from task 3", nil
}
