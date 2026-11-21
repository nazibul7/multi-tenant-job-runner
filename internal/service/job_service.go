package service

import (
    "crypto/rand"
    "encoding/hex"
    "time"

    "github.com/nazibul7/multi-tenant-job-runner/internal/job"
    "github.com/nazibul7/multi-tenant-job-runner/internal/store"
)

type JobService struct {
    tenantStore *store.TenantStore
    jobStore    *store.JobStore
}

func NewJobService(
    tenantStore *store.TenantStore,
    jobStore *store.JobStore,
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

    // Generate job ID.
    id, err := generateJobID()
    if err != nil {
        return err
    }

    j.ID = id
    j.TenantID = tenantID

    // Validate job.
    if err := job.Validate(*j); err != nil {
        return err
    }

    // Initialize job.
    j.Status = job.JobQueued
    j.CreatedAt = time.Now().UTC()

    // Initialize steps.
    for _, step := range j.Steps {
        step.Status = job.StepQueued
        step.Attempts = 0
        step.Error = ""
    }

    // Save job.
    return s.jobStore.Create(j)
}

func (s *JobService) GetJob(tenantID, jobID string) (*job.Job, error) {
    return s.jobStore.GetForTenant(jobID, tenantID)
}

func generateJobID() (string, error) {
    b := make([]byte, 16)

    if _, err := rand.Read(b); err != nil {
        return "", err
    }

    return hex.EncodeToString(b), nil
}