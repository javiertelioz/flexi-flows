package uses_cases

// ConditionFunc
// @type: task
// @description Condition for conditional node.
//
// @input data (interface{}): Input data.
//
// @output bool: Returns true if condition is met, false otherwise.
func ConditionFunc(data interface{}) bool {
	result, ok := data.(string)
	return ok && result == "result from task 1"
}
