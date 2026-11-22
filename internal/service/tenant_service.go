package service

import (
	"github.com/nazibul7/multi-tenant-job-runner/internal/job"
	"github.com/nazibul7/multi-tenant-job-runner/internal/store"
)

type TenantUsage struct {
	TenantID  string `json:"tenant_id"`
	Budget    int    `json:"budget"`
	Spent     int    `json:"spent"`
	Remaining int    `json:"remaining"`
	Reserved  int    `json:"reserved"`
	Queued    int    `json:"queued"`
	Running   int    `json:"running"`
	Succeeded int    `json:"succeeded"`
	Failed    int    `json:"failed"`
}

type TenantService struct {
	tenantStore *store.TenantStore
	jobStore    *store.JobStore
}

func NewTenantService(
	tenantStore *store.TenantStore,
	jobStore *store.JobStore,
) *TenantService {
	return &TenantService{
		tenantStore: tenantStore,
		jobStore:    jobStore,
	}
}

func (s *TenantService) GetUsage(tenantID string) (*TenantUsage, error) {
	t, err := s.tenantStore.Get(tenantID)
	if err != nil {
		return nil, err
	}

	jobs := s.jobStore.ListByTenant(tenantID)

	usage := &TenantUsage{
		TenantID:  t.ID,
		Budget:    t.Budget,
		Spent:     t.Spent,
		Remaining: t.Budget - t.Spent,
		Reserved:  0,
	}

	for _, j := range jobs {
		switch j.Status {
		case job.JobQueued:
			usage.Queued++
		case job.JobRunning:
			usage.Running++
		case job.JobSucceeded:
			usage.Succeeded++
		case job.JobFailed:
			usage.Failed++
		}
	}

	return usage, nil
}
