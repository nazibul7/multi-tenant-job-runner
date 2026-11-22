package job

import "errors"

var (
	ErrEmptySteps         = errors.New("job must contain steps")
	ErrEmptyStepID        = errors.New("step ID cannot be empty")
	ErrDuplicateStepID    = errors.New("duplicate step ID")
	ErrInvalidJob         = errors.New("invalid job")
	ErrNilStep            = errors.New("step cannot be nil")
	ErrUnknownDependency  = errors.New("dependency does not exist")
	ErrSelfDependency     = errors.New("step cannot depend on itself")
	ErrCircularDependency = errors.New("circular dependency detected")
)

func Validate(j Job) error {
	if j.ID == "" || j.TenantID == "" {
		return ErrInvalidJob
	}

	if len(j.Steps) == 0 {
		return ErrEmptySteps
	}

	stepIDs := make(map[string]bool)

	// First, validate step IDs.
	for _, step := range j.Steps {
		if step == nil {
			return ErrNilStep
		}

		if step.ID == "" {
			return ErrEmptyStepID
		}

		if stepIDs[step.ID] {
			return ErrDuplicateStepID
		}

		stepIDs[step.ID] = true
	}

	// Then, validate dependencies.
	graph := make(map[string][]string)

	for _, step := range j.Steps {
		for _, dependency := range step.DependsOn {
			if dependency == step.ID {
				return ErrSelfDependency
			}

			if !stepIDs[dependency] {
				return ErrUnknownDependency
			}

			graph[step.ID] = append(graph[step.ID], dependency)
		}
	}

	// Detect circular dependencies using DFS.
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var visit func(string) bool

	visit = func(id string) bool {
		if inStack[id] {
			return false
		}

		if visited[id] {
			return true
		}

		visited[id] = true
		inStack[id] = true

		for _, dependency := range graph[id] {
			if !visit(dependency) {
				return false
			}
		}

		inStack[id] = false
		return true
	}

	for _, step := range j.Steps {
		if !visit(step.ID) {
			return ErrCircularDependency
		}
	}

	return nil
}
