package fixtures

import "fmt"

// Calculate performs a complex calculation
func Calculate(input int) int {
	step1 := Process(input)
	step2 := Transform(step1)
	return Finalize(step2)
}

// Process validates and processes input
func Process(value int) int {
	if value < 0 {
		return 0
	}
	return value * 2
}

// Transform applies transformation logic
func Transform(value int) int {
	return value + 10
}

// Finalize completes the calculation
func Finalize(value int) int {
	result := value * 3
	fmt.Printf("Final result: %d\n", result)
	return result
}

// Helper is a utility function
func Helper(msg string) string {
	return fmt.Sprintf("Helper: %s", msg)
}
