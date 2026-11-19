package job

import "time"

type StepStatus string

const (
	StepQueued    StepStatus = "queued"
	StepRunning   StepStatus = "running"
	StepSucceeded StepStatus = "succeeded"
	StepFailed    StepStatus = "failed"
	StepBlocked   StepStatus = "blocked"
)

type JobStatus string

const (
	JobQueued    JobStatus = "queued"
	JobRunning   JobStatus = "running"
	JobSucceeded JobStatus = "succeeded"
	JobFailed    JobStatus = "failed"
)

type Job struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Steps     []*Step   `json:"steps"`
	Status    JobStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Step struct {
	ID          string     `json:"id"`
	DependsOn   []string   `json:"depends_on"`
	DurationMS  int        `json:"duration_ms"`
	Cost        int        `json:"cost"`
	FailTimes   int        `json:"fail_times"`
	MaxAttempts int        `json:"max_attempts"`
	Attempts    int        `json:"attempts"`
	Status      StepStatus `json:"status"`
	Error       string     `json:"error,omitempty"`
}
