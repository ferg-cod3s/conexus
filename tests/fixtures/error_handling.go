package fixtures

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
	ErrTimeout      = errors.New("operation timeout")
)

// Divide performs division with error handling
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrInvalidInput
	}
	return a / b, nil
}

// Fetch retrieves data with multiple error conditions
func Fetch(id int) (string, error) {
	if id < 0 {
		return "", ErrInvalidInput
	}
	if id > 100 {
		return "", ErrNotFound
	}
	return fmt.Sprintf("Data-%d", id), nil
}

// ProcessWithRetry attempts operation with retry logic
func ProcessWithRetry(data string) error {
	if data == "" {
		return ErrInvalidInput
	}

	err := attemptProcess(data)
	if err != nil {
		// Retry once
		err = attemptProcess(data)
		if err != nil {
			return fmt.Errorf("process failed after retry: %w", err)
		}
	}
	return nil
}

func attemptProcess(data string) error {
	if len(data) < 3 {
		return errors.New("data too short")
	}
	return nil
}

// ValidateAndProcess validates input then processes
func ValidateAndProcess(input string) (string, error) {
	if input == "" {
		return "", ErrInvalidInput
	}

	validated := validate(input)
	if !validated {
		return "", errors.New("validation failed")
	}

	result, err := process(input)
	if err != nil {
		return "", fmt.Errorf("processing error: %w", err)
	}

	return result, nil
}

func validate(s string) bool {
	return len(s) > 0
}

func process(s string) (string, error) {
	return "processed: " + s, nil
}
