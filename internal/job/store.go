package job

import (
	"errors"
	"sync"
)

var (
	ErrJobNotFound = errors.New("job not found")
	ErrJobExists   = errors.New("job already exists")
	ErrInvalidJob  = errors.New("invalid job")
)

type Store struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewStore() *Store {
	return &Store{
		jobs: make(map[string]*Job),
	}
}

func (s *Store) Create(j *Job) error {
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

func (s *Store) Get(jobID string) (*Job, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    j, exists := s.jobs[jobID]
    if !exists {
        return nil, ErrJobNotFound
    }

    return j, nil
}
