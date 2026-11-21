package service

import (
	"errors"

	"github.com/nazibul7/multi-tenant-job-runner/internal/job"
	"github.com/nazibul7/multi-tenant-job-runner/internal/tenant"
)

var ErrTenantMismatch = errors.New("job does not belong to submitting tenant")

type JobService struct {
    tenantStore *tenant.Store
    jobStore    *job.Store
}

func NewJobService(
    tenantStore *tenant.Store,
    jobStore *job.Store,
) *JobService {
    return &JobService{
        tenantStore: tenantStore,
        jobStore:    jobStore,
    }
}

func (s *JobService) SubmitJob(tenantID string, j *job.Job) error {
    if j == nil {
        return job.ErrInvalidJob
    }

    // Check whether tenant exists.
    _, err := s.tenantStore.Get(tenantID)
    if err != nil {
        return err
    }

    // Check job ownership.
    if j.TenantID != tenantID {
        return ErrTenantMismatch
    }

    // Validate job.
    if err := job.Validate(*j); err != nil {
        return err
    }

    // Save job.
    return s.jobStore.Create(j)
}