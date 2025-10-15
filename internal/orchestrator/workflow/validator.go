package workflow

import (
	"fmt"
)

// Validator validates workflow structures
type Validator struct {
}

// NewValidator creates a new workflow validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks if a workflow is valid
func (v *Validator) Validate(wf *Workflow) error {
	if wf == nil {
		return fmt.Errorf("workflow is nil")
	}

	if wf.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	if len(wf.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	// Validate each step
	stepIDs := make(map[string]bool)
	for i, step := range wf.Steps {
		if err := v.validateStep(step, i); err != nil {
			return err
		}

		// Check for duplicate step IDs
		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %s", step.ID)
		}
		stepIDs[step.ID] = true
	}

	// Validate dependencies for parallel workflows
	if wf.Mode == ParallelMode {
		if err := v.validateDependencies(wf); err != nil {
			return err
		}
	}

	return nil
}

// validateStep validates a single workflow step
func (v *Validator) validateStep(step *Step, index int) error {
	if step == nil {
		return fmt.Errorf("step %d is nil", index)
	}

	if step.ID == "" {
		return fmt.Errorf("step %d: ID is required", index)
	}

	if step.Agent == "" {
		return fmt.Errorf("step %d (%s): agent is required", index, step.ID)
	}

	return nil
}

// validateDependencies validates that all dependencies exist
func (v *Validator) validateDependencies(wf *Workflow) error {
	stepIDs := make(map[string]bool)
	for _, step := range wf.Steps {
		stepIDs[step.ID] = true
	}

	for _, step := range wf.Steps {
		for _, depID := range step.Dependencies {
			if !stepIDs[depID] {
				return fmt.Errorf("step %s has unknown dependency: %s", step.ID, depID)
			}
		}
	}

	// Check for circular dependencies
	if err := v.checkCircularDependencies(wf); err != nil {
		return err
	}

	return nil
}

// checkCircularDependencies detects circular dependencies in the workflow
func (v *Validator) checkCircularDependencies(wf *Workflow) error {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	// Build dependency graph
	graph := make(map[string][]string)
	for _, step := range wf.Steps {
		graph[step.ID] = step.Dependencies
	}

	// DFS to detect cycles
	var hasCycle func(string) bool
	hasCycle = func(stepID string) bool {
		visited[stepID] = true
		recursionStack[stepID] = true

		for _, depID := range graph[stepID] {
			if !visited[depID] {
				if hasCycle(depID) {
					return true
				}
			} else if recursionStack[depID] {
				return true
			}
		}

		recursionStack[stepID] = false
		return false
	}

	for _, step := range wf.Steps {
		if !visited[step.ID] {
			if hasCycle(step.ID) {
				return fmt.Errorf("circular dependency detected involving step: %s", step.ID)
			}
		}
	}

	return nil
}
