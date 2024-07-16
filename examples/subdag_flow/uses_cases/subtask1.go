package uses_cases

import "fmt"

// SubTask1
// @type: task
// @description Execute SubTask 1.
//
// @input data (interface{}): Input data.
//
// @output string: Returns result from subtask 1.
// @output error: Returns an error if occurs.
func SubTask1(data interface{}) (string, error) {
	fmt.Println("Executing SubTask 1")
	return "result from subtask 1", nil
}
