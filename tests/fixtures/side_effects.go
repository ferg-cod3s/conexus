package fixtures

import (
	"fmt"
	"log"
	"os"
)

// LogOperation performs an operation and logs it
func LogOperation(name string, value int) int {
	log.Printf("Starting operation: %s with value %d", name, value)
	result := value * 2
	log.Printf("Operation %s completed with result: %d", name, result)
	return result
}

// WriteToFile writes data to a file
func WriteToFile(filename string, data []byte) error {
	log.Printf("Writing %d bytes to file: %s", len(data), filename)
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Printf("Error writing to file: %v", err)
		return err
	}
	log.Printf("Successfully wrote to file: %s", filename)
	return nil
}

// ReadFromFile reads data from a file
func ReadFromFile(filename string) ([]byte, error) {
	log.Printf("Reading from file: %s", filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return nil, err
	}
	log.Printf("Successfully read %d bytes from file: %s", len(data), filename)
	return data, nil
}

// ProcessWithSideEffects demonstrates multiple side effects
func ProcessWithSideEffects(input string) string {
	// Log side effect
	fmt.Println("Processing input:", input)

	// Metric side effect (simulated)
	recordMetric("process_count", 1)

	// State mutation
	result := transform(input)

	// Another log
	fmt.Println("Processing complete:", result)

	return result
}

func recordMetric(name string, value int) {
	fmt.Printf("METRIC: %s = %d\n", name, value)
}

func transform(s string) string {
	return "transformed: " + s
}

// NotifyUser sends a notification
func NotifyUser(userID int, message string) error {
	log.Printf("Sending notification to user %d: %s", userID, message)
	// Simulated HTTP call side effect
	fmt.Printf("HTTP POST /notify/%d {message: %s}\n", userID, message)
	return nil
}
