package fixtures

import "fmt"

// Calculator represents a stateful calculator
type Calculator struct {
	Total int
	Count int
}

// NewCalculator creates a new Calculator instance
func NewCalculator() *Calculator {
	return &Calculator{
		Total: 0,
		Count: 0,
	}
}

// Add adds a value to the total
func (c *Calculator) Add(value int) {
	c.Total += value
	c.Count++
}

// Subtract subtracts a value from the total
func (c *Calculator) Subtract(value int) {
	c.Total -= value
	c.Count++
}

// Average calculates the average
func (c *Calculator) Average() float64 {
	if c.Count == 0 {
		return 0.0
	}
	return float64(c.Total) / float64(c.Count)
}

// Reset resets the calculator state
func (c *Calculator) Reset() {
	c.Total = 0
	c.Count = 0
}

// Display prints the current state
func (c *Calculator) Display() {
	fmt.Printf("Total: %d, Count: %d, Average: %.2f\n",
		c.Total, c.Count, c.Average())
}
