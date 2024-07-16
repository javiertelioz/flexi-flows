package uses_cases

import "fmt"

// MainTask
// @type: task
// @description Execute Main Task.
//
// @input data (interface{}): Input data.
//
// @output interface{}: Returns result from Main Task.
// @output error: Returns an error if occurs.
func MainTask(data interface{}) (interface{}, error) {
	fmt.Println("Executing Main Task")
	return "result from Main Task", nil
}
