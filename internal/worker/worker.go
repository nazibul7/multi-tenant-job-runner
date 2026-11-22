package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nazibul7/multi-tenant-job-runner/internal/job"
	scheduler "github.com/nazibul7/multi-tenant-job-runner/internal/sheduler"
	"github.com/nazibul7/multi-tenant-job-runner/internal/store"
)

type Pool struct {
	jobs        <-chan *job.Job
	jobStore    *store.JobStore
	scheduler   *scheduler.Scheduler
	workerCount int
}

func NewPool(
	workerCount int,
	jobs <-chan *job.Job,
	jobStore *store.JobStore,
	scheduler *scheduler.Scheduler,
) *Pool {
	if workerCount <= 0 {
		workerCount = 1
	}

	return &Pool{
		jobs:        jobs,
		jobStore:    jobStore,
		scheduler:   scheduler,
		workerCount: workerCount,
	}
}

// Run starts the worker goroutines.
func (p *Pool) Run(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			p.startWorker(ctx)
		}()
	}

	wg.Wait()
}

// Each worker waits for jobs from the scheduler.
func (p *Pool) startWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case j, ok := <-p.jobs:
			if !ok {
				return
			}

			if j != nil {
				p.executeJob(ctx, j)
			}
		}
	}
}

// executeJob executes all steps of a job.

func (p *Pool) executeJob(ctx context.Context, j *job.Job) {
	defer p.scheduler.JobFinished(j.TenantID)

	completed := make(map[string]bool)
	processed := make(map[string]bool)

	remaining := len(j.Steps)

	for remaining > 0 {
		progress := false

		for _, step := range j.Steps {
			if processed[step.ID] {
				continue
			}

			// Check whether all dependencies succeeded.
			dependenciesReady := true

			for _, dependency := range step.DependsOn {
				if !completed[dependency] {
					dependenciesReady = false
					break
				}
			}

			if !dependenciesReady {
				continue
			}

			progress = true

			if err := p.executeStep(ctx, j.ID, step); err != nil {

				processed[step.ID] = true // Don't overwrite the failed step.

				for _, remainingStep := range j.Steps {
					if remainingStep != nil && !processed[remainingStep.ID] {
						_ = p.jobStore.UpdateStep(
							j.ID,
							remainingStep.ID,
							job.StepBlocked,
							0,
							"previous step failed",
						)
					}
				}

				p.jobStore.SetStatus(j.ID, job.JobFailed)
				return
			}

			completed[step.ID] = true
			processed[step.ID] = true
			remaining--
		}

		// No step was ready to execute.
		if !progress {
			p.jobStore.SetStatus(j.ID, job.JobFailed)
			return
		}
	}

	p.jobStore.SetStatus(j.ID, job.JobSucceeded)
}

// executeStep runs a step and retries simulated failures.
func (p *Pool) executeStep(
	ctx context.Context,
	jobID string,
	step *job.Step,
) error {
	maxAttempts := step.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			_ = p.jobStore.UpdateStep(
				jobID,
				step.ID,
				job.StepFailed,
				attempt-1,
				err.Error(),
			)
			return err
		}

		_ = p.jobStore.UpdateStep(
			jobID,
			step.ID,
			job.StepRunning,
			attempt,
			"",
		)

		// Simulate step execution time.
		timer := time.NewTimer(
			time.Duration(step.DurationMS) * time.Millisecond,
		)

		select {
		case <-ctx.Done():
			timer.Stop()

			_ = p.jobStore.UpdateStep(
				jobID,
				step.ID,
				job.StepFailed,
				attempt,
				ctx.Err().Error(),
			)

			return ctx.Err()

		case <-timer.C:
		}

		// FailTimes simulates the first N attempts failing.
		if attempt <= step.FailTimes {
			err := errors.New("simulated step failure")

			_ = p.jobStore.UpdateStep(
				jobID,
				step.ID,
				job.StepFailed,
				attempt,
				err.Error(),
			)

			if attempt == maxAttempts {
				return fmt.Errorf(
					"step %s failed after %d attempts",
					step.ID,
					attempt,
				)
			}

			continue
		}

		_ = p.jobStore.UpdateStep(
			jobID,
			step.ID,
			job.StepSucceeded,
			attempt,
			"",
		)

		return nil
	}

	return fmt.Errorf("step %s failed", step.ID)
}
