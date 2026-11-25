package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/nazibul7/multi-tenant-job-runner/internal/job"
	"github.com/nazibul7/multi-tenant-job-runner/internal/store"
)

var (
	ErrMissingIdempotencyKey = errors.New("idempotency key is required")
	ErrIdempotencyConflict   = errors.New("idempotency key reused with different request")
)

type JobService struct {
	tenantStore      *store.TenantStore
	jobStore         *store.JobStore
	idempotencyStore *store.IdempotencyStore
}

func NewJobService(
	tenantStore *store.TenantStore,
	jobStore *store.JobStore,
	idempotencyStore *store.IdempotencyStore,
) *JobService {
	return &JobService{
		tenantStore:      tenantStore,
		jobStore:         jobStore,
		idempotencyStore: idempotencyStore,
	}
}

func (s *JobService) SubmitJob(tenantID string, idempotencyKey string, j *job.Job) error {
	if idempotencyKey == "" {
		return ErrMissingIdempotencyKey
	}

	if j == nil {
		return job.ErrInvalidJob
	}

	requestHash, err := hashRequest(j)
	if err != nil {
		return err
	}

	// Scope the key to its tenant.
	scopedKey := tenantID + "\x00" + idempotencyKey

	// Check for an existing submission.
	record, exists := s.idempotencyStore.Get(scopedKey)
	if exists {
		if record.RequestHash != requestHash {
			return ErrIdempotencyConflict
		}

		// Return the original job.
		existingJob, err := s.jobStore.GetForTenant(record.JobID, tenantID)
		if err != nil {
			return err
		}

		*j = *existingJob
		return nil
	}

	// Check whether tenant exists.
	_, err = s.tenantStore.Get(tenantID)
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

	// Calculate total cost of all steps.
	totalCost := 0

	for _, step := range j.Steps {
		if step.Cost < 0 {
			return errors.New("step cost cannot be negative")
		}

		totalCost += step.Cost
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

	// Charge tenant budget.
	if err := s.tenantStore.TryCharge(tenantID, totalCost); err != nil {
		return err
	}

	// Save job.
	if err := s.jobStore.Create(j); err != nil {
		// Refund if saving the job fails.
		s.tenantStore.Refund(tenantID, totalCost)
		return err
	}

	// Save idempotency record only after successful job creation.
	s.idempotencyStore.Save(scopedKey, store.IdempotencyRecord{
		RequestHash: requestHash,
		JobID:       j.ID,
	})

	return nil
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
