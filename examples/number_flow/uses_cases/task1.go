package uses_cases

import "fmt"

// Task1
// @type: task
// @description Execute Task 1.
//
// @input string: Input data.
//
// @output string: Returns result from task 1.
// @output error: Returns an error if occurs.
func Task1(start string) (string, error) {
	fmt.Println("Executing Task 1")
	fmt.Println(start)
	return "result from task 1", nil
}

// AfterTask1
// @type: post-hook
// @after Task1
// @description Execute after Task 1.
func AfterTask1() {
	fmt.Println("Finished Task 1")
}
