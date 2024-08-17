package uses_cases

import "fmt"

// Branch2
// @type: task
// @description Execute Branch2.
//
// @input data (interface{}): Input data.
//
// @output string: Returns result from Branch2.
// @output error: Returns an error if occurs.
func Branch2(data interface{}) (string, error) {
	fmt.Println("Executing Branch2")
	return "result from Branch2", nil
}
