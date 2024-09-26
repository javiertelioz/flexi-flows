package use_cases

import (
	"fmt"
	"log"
	"math"
)

// IsPrimeFunc
// @type: task
// @description checks if a given integer is a prime number.
//
// @input data (int): The integer to be checked for primality.
//
// @output bool: Returns true if the input integer is a prime number, false otherwise.
// @output error: Returns an error if the input integer is less than 1.
func IsPrimeFunc(data any) (any, error) {
	number, ok := data.(int)
	if !ok {
		return nil, fmt.Errorf("IsPrimeFunc expects data of type int, got %T", data)
	}

	if number <= 1 {
		return false, nil
	}

	for i := 2; i <= int(math.Sqrt(float64(number))); i++ {
		if number%i == 0 {
			return false, nil
		}
	}
	return true, nil
}

// BeforeIsPrime
// @type: pre-hook
// @before IsPrimeFunc
// @description execute before IsPrimeFunc.
//
// @input data (int): The integer to be checked for primality.
//
// @output int: Returns integer is a prime number.
// @output error: Returns an error if occurs.
func BeforeIsPrime(data any) (any, error) {
	log.Printf("Before isPrime: received data %v", data)
	return data, nil
}

// AfterIsPrime
// @type: post-hook
// @after IsPrimeFunc
// @description execute after IsPrimeFunc.
//
// @input result (bool): The result of the IsPrimeFunc.
//
// @output bool: Returns the result after post-processing.
// @output error: Returns an error if occurs.
func AfterIsPrime(result any) (any, error) {
	isPrime, ok := result.(bool)
	if !ok {
		return nil, fmt.Errorf("AfterIsPrime expects result of type bool, got %T", result)
	}
	log.Printf("After isPrime: result is %v", isPrime)
	return result, nil
}
