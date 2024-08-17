package use_cases

import (
	"fmt"
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
func IsPrimeFunc(data int) (bool, error) {
	if data <= 1 {
		return false, nil
	}
	for i := 2; i <= int(math.Sqrt(float64(data))); i++ {
		if data%i == 0 {
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
func BeforeIsPrime(data int) (int, error) {
	fmt.Println("Before isPrime: received data", data)
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
func AfterIsPrime(result bool) (bool, error) {
	fmt.Println("After isPrime: result is", result)
	return result, nil
}
