package scheduler

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/nazibul7/multi-tenant-job-runner/internal/job"
	"github.com/nazibul7/multi-tenant-job-runner/internal/store"
)

type Scheduler struct {
	tenantStore *store.TenantStore
	jobStore    *store.JobStore
	jobs        chan *job.Job

	// Index of the tenant whose turn is next.
	nextTenant int
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

// schedule dispatches jobs using round-robin tenant selection.
func (s *Scheduler) schedule(ctx context.Context) {
	tenants := s.tenantStore.ListAll()

	if len(tenants) == 0 {
		return
	}

	// Keep tenant order stable.
	sort.Slice(tenants, func(i, j int) bool {
		return tenants[i].ID < tenants[j].ID
	})

	tenantCount := len(tenants)

	for {
		// Start from the tenant whose turn is next.
		start := s.nextTenant % tenantCount

		dispatched := false

		// Give every tenant one opportunity in this round.
		for i := 0; i < tenantCount; i++ {
			index := (start + i) % tenantCount
			t := tenants[index]

			// Check tenant concurrency limit.
			if !s.tenantStore.TryStartJob(t.ID) {
				continue
			}

			// Claim one queued job for this tenant.
			j := s.jobStore.ClaimNextQueued(t.ID)

			if j == nil {
				s.tenantStore.FinishJob(t.ID)
				continue
			}

			// Dispatch job to buffered worker channel.
			select {
			case s.jobs <- j:
				dispatched = true

				// Next dispatch starts with the next tenant.
				s.nextTenant = (index + 1) % tenantCount

				log.Printf("DISPATCH tenant=%s job=%s nextTenant=%s",
					t.ID,
					j.ID,
					tenants[s.nextTenant].ID,
				)

			case <-ctx.Done():
				s.jobStore.SetStatus(j.ID, job.JobQueued)
				s.tenantStore.FinishJob(t.ID)
				return
			}
		}

		// Stop if no tenant could dispatch a job.
		if !dispatched {
			return
		}
	}
}

// JobFinished releases the tenant's running slot.
func (s *Scheduler) JobFinished(tenantID string) {
	s.tenantStore.FinishJob(tenantID)
}
