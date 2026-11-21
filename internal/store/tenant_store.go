package store

import (
	"errors"
	"sync"

	"github.com/nazibul7/multi-tenant-job-runner/internal/tenant"
)

var (
	ErrTenantExists   = errors.New("tenant already exists")
	ErrTenantNotFound = errors.New("tenant not found")
	ErrInvalidTenant  = errors.New("invalid tenant configuration")
)

type TenantStore struct {
	mu      sync.RWMutex
	tenants map[string]*tenant.Tenant
}

func NewTenantStore() *TenantStore {
	return &TenantStore{
		tenants: make(map[string]*tenant.Tenant),
	}
}

func (s *TenantStore) Create(t tenant.Tenant) error {
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

func (s *TenantStore) Get(id string) (*tenant.Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, exists := s.tenants[id]
	if !exists {
		return nil, ErrTenantNotFound
	}

	// Return a copy to prevent external modification.
	tenantCopy := *t
	return &tenantCopy, nil
}