package tenant

import (
	"errors"
	"sync"
)

var (
	ErrTenantExists   = errors.New("tenant already exists")
	ErrTenantNotFound = errors.New("tenant not found")
	ErrInvalidTenant  = errors.New("invalid tenant configuration")
)

type Store struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
}

func NewStore() *Store {
	return &Store{
		tenants: make(map[string]*Tenant),
	}
}

func (s *Store) Create(t Tenant) error {
	if t.ID == "" || t.Budget <= 0 || t.MaxConcurrency <= 0 {
		return ErrInvalidTenant
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[t.ID]; exists {
		return ErrTenantExists
	}

	s.tenants[t.ID] = &t

	return nil
}

func (s *Store) Get(id string) (*Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, exists := s.tenants[id]
	if !exists {
		return nil, ErrTenantNotFound
	}

	return t, nil
}
