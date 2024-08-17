package uses_cases

import "fmt"

// FalseBranch
// @type: task
// @description Execute FalseBranch if condition is not met.
//
// @input data (interface{}): Input data.
//
// @output string: Returns result from FalseBranch.
// @output error: Returns an error if occurs.
func FalseBranch(data interface{}) (string, error) {
	fmt.Println("Executing FalseBranch")
	return "result from FalseBranch", nil
}
