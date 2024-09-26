package use_cases

import (
	"fmt"
	"log"
)

// SquareFunc
// @type: task
// @description calculates the square of a given integer.
//
// @input data (int): The integer to be squared.
//
// @output int: Returns the square of the input integer.
// @output error: Returns an error if occurs.
func SquareFunc(data any) (any, error) {
	number, ok := data.(int)
	if !ok {
		return nil, fmt.Errorf("SquareFunc expects data of type int, got %T", data)
	}
	return number * number, nil
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
func BeforeSquare(data any) (any, error) {
	number, ok := data.(int)
	if !ok {
		return nil, fmt.Errorf("BeforeSquare expects data of type int, got %T", data)
	}
	log.Printf("Before square: received data %d", number)
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
func AfterSquare(result any) (any, error) {
	square, ok := result.(int)
	if !ok {
		return nil, fmt.Errorf("AfterSquare expects result of type int, got %T", result)
	}
	log.Printf("After square: result is %d", square)
	return result, nil
}
