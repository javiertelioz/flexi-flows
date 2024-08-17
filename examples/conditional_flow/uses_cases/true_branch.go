package uses_cases

import "fmt"

// TrueBranch
// @type: task
// @description Execute TrueBranch if condition is met.
//
// @input data (interface{}): Input data.
//
// @output string: Returns result from TrueBranch.
// @output error: Returns an error if occurs.
func TrueBranch(data interface{}) (string, error) {
	fmt.Println("Executing TrueBranch")
	return "result from TrueBranch", nil
}
