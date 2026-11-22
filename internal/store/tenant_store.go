package store

import (
	"errors"
	"sync"

	"github.com/nazibul7/multi-tenant-job-runner/internal/tenant"
)

var (
	ErrTenantExists       = errors.New("tenant already exists")
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrInvalidTenant      = errors.New("invalid tenant configuration")
	ErrInsufficientBudget = errors.New("insufficient tenant budget")
	ErrInvalidAmount      = errors.New("invalid step amount")
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

func (s *TenantStore) ListAll() []*tenant.Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*tenant.Tenant

	for _, t := range s.tenants {
		tenantCopy := *t
		result = append(result, &tenantCopy)
	}

	return result
}

// TryStartJob reserves one running slot if capacity is available.
func (s *TenantStore) TryStartJob(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tenants[id]
	if !exists {
		return false
	}

	if t.Running >= t.MaxConcurrency {
		return false
	}

	t.Running++
	return true
}

// FinishJob releases one running slot.
func (s *TenantStore) FinishJob(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tenants[id]
	if !exists {
		return
	}

	if t.Running > 0 {
		t.Running--
	}
}

func (s *TenantStore) TryCharge(id string, amount int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tenants[id]
	if !exists {
		return ErrTenantNotFound
	}

	if amount < 0 {
		return ErrInvalidAmount
	}

	remaining := t.Budget - t.Spent

	if amount > remaining {
		return ErrInsufficientBudget
	}

	t.Spent += amount

	return nil
}

func (s *TenantStore) Refund(id string, amount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tenants[id]
	if !exists {
		return
	}

	if amount <= 0 {
		return
	}

	t.Spent -= amount

	if t.Spent < 0 {
		t.Spent = 0
	}
}
