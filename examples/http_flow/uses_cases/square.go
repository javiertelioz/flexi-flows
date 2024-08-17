package use_cases

import "fmt"

// SquareFunc
// @type: task
// @description calculates the square of a given integer.
//
// @input data (int): The integer to be squared.
//
// @output int: Returns the square of the input integer.
// @output error: Returns an error if occurs.
func SquareFunc(data int) (int, error) {
	return data * data, nil
}

// BeforeSquare
// @type: pre-hook
// @before SquareFunc
// @description execute before SquareFunc.
//
// @input data (int): The integer to be squared.
//
// @output int: Returns integer to be squared.
// @output error: Returns an error if occurs.
func BeforeSquare(data int) (int, error) {
	fmt.Println("Before square: received data", data)
	return data, nil
}

// AfterSquare
// @type: post-hook
// @after SquareFunc
// @description execute after SquareFunc.
//
// @input result (int): The result of the SquareFunc.
//
// @output int: Returns the result after post-processing.
// @output error: Returns an error if occurs.
func AfterSquare(result int) (int, error) {
	fmt.Println("After square: result is", result)
	return result, nil
}
