package store

import (
	"errors"
	"sync"

	"github.com/nazibul7/multi-tenant-job-runner/internal/job"
)

var (
	ErrJobNotFound = errors.New("job not found")
	ErrJobExists   = errors.New("job already exists")
	ErrInvalidJob  = errors.New("invalid job")
)

type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*job.Job
}

func NewJobStore() *JobStore {
	return &JobStore{
		jobs: make(map[string]*job.Job),
	}
}

func (s *JobStore) Create(j *job.Job) error {
	if j == nil || j.ID == "" || j.TenantID == "" {
		return ErrInvalidJob
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[j.ID]; exists {
		return ErrJobExists
	}

	s.jobs[j.ID] = j

	return nil
}

func (s *JobStore) GetForTenant(jobID, tenantID string) (*job.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	j, exists := s.jobs[jobID]
	if !exists || j.TenantID != tenantID {
		return nil, ErrJobNotFound
	}

	return j, nil
}

func (s *JobStore) ListByTenant(tenantID string) []*job.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*job.Job

	for _, j := range s.jobs {
		if j.TenantID == tenantID {
			result = append(result, j)
		}
	}

	return result
}