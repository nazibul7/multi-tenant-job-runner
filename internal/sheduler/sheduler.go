package scheduler

import (
	"context"
	"time"

	"github.com/nazibul7/multi-tenant-job-runner/internal/job"
	"github.com/nazibul7/multi-tenant-job-runner/internal/store"
)

type Scheduler struct {
	tenantStore *store.TenantStore
	jobStore    *store.JobStore
	jobs        chan *job.Job
}

func NewScheduler(
	tenantStore *store.TenantStore,
	jobStore *store.JobStore,
) *Scheduler {
	return &Scheduler{
		tenantStore: tenantStore,
		jobStore:    jobStore,
		jobs:        make(chan *job.Job, 100),
	}
}

// Jobs returns a read-only channel for workers.
func (s *Scheduler) Jobs() <-chan *job.Job {
	return s.jobs
}

// Run starts the scheduler loop.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	defer close(s.jobs)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			s.schedule(ctx)
		}
	}
}

// schedule finds queued jobs and dispatches eligible jobs.
func (s *Scheduler) schedule(ctx context.Context) {
	tenants := s.tenantStore.ListAll()

	for _, t := range tenants {
		for {
			// Check tenant concurrency limit.
			if !s.tenantStore.TryStartJob(t.ID) {
				break
			}

			// Claim one queued job.
			j := s.jobStore.ClaimNextQueued(t.ID)

			if j == nil {
				s.tenantStore.FinishJob(t.ID)
				break
			}

			// Dispatch job to the worker channel.
			select {
			case s.jobs <- j:
				// Successfully dispatched.

			case <-ctx.Done():
				s.jobStore.SetStatus(j.ID, job.JobQueued)
				s.tenantStore.FinishJob(t.ID)
				return
			}
		}
	}
}

// JobFinished releases the tenant's running slot.
func (s *Scheduler) JobFinished(tenantID string) {
	s.tenantStore.FinishJob(tenantID)
}
