package job

import "errors"

var (
	ErrEmptySteps      = errors.New("job must contain steps")
	ErrEmptyStepID     = errors.New("step ID cannot be empty")
	ErrDuplicateStepID = errors.New("duplicate step ID")
	ErrInvalidJob      = errors.New("invalid job")
	ErrNilStep         = errors.New("step cannot be nil")
)

func Validate(j Job) error {
	// Check Job ID and Tenant ID
	if j.ID == "" || j.TenantID == "" {
		return ErrInvalidJob
	}

	// Check whether job has steps
	if len(j.Steps) == 0 {
		return ErrEmptySteps
	}

	// Check step IDs
	stepIDs := make(map[string]bool)

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

	return nil
}
