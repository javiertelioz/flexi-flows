package uses_cases

import "fmt"

// IterateFunc
// @type: task
// @description Iterate over a collection.
//
// @input item (interface{}): Input item.
//
// @output interface{}: Returns nil.
// @output error: Returns an error if occurs.
func IterateFunc(item interface{}) (interface{}, error) {
	fmt.Printf("Processing item: %v\n", item)
	return nil, nil
}
