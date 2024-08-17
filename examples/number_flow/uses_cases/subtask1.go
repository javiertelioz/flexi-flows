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

// BeforeSubTask1
// @type: pre-hook
// @before SubTask1
// @description Execute before SubTask 1.
func BeforeSubTask1() {
	fmt.Println("Starting SubTask 1")
}

// AfterSubTask1
// @type: post-hook
// @after SubTask1
// @description Execute after SubTask 1.
func AfterSubTask1() {
	fmt.Println("Finished SubTask 1")
}
