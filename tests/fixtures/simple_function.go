package fixtures

import "fmt"

// Add performs addition of two integers
func Add(a, b int) int {
	return a + b
}

// Greet prints a greeting message
func Greet(name string) {
	fmt.Printf("Hello, %s!\n", name)
}

// Multiply returns the product of two numbers
func Multiply(x, y int) int {
	result := x * y
	return result
}
