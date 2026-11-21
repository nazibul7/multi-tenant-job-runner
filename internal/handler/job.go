package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nazibul7/multi-tenant-job-runner/internal/job"
	"github.com/nazibul7/multi-tenant-job-runner/internal/service"
	"github.com/nazibul7/multi-tenant-job-runner/internal/tenant"
)

type JobHandler struct {
    jobService *service.JobService
}

func NewJobHandler(jobService *service.JobService) *JobHandler {
    return &JobHandler{
        jobService: jobService,
    }
}

func (h *JobHandler) SubmitJob(w http.ResponseWriter, r *http.Request) {
    tenantID := r.PathValue("tenantID")

    if tenantID == "" {
        http.Error(w, "tenant ID is required", http.StatusBadRequest)
        return
    }

    var j job.Job

    // Decode request body.
	const maxRequestBodySize = 1024 * 1024
    r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

    if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    // Tenant ID comes from the URL, not the request body.
    j.TenantID = tenantID

    // Submit job to service.
    err := h.jobService.SubmitJob(tenantID, &j)
    if err != nil {
        switch {
        case errors.Is(err, tenant.ErrTenantNotFound):
            http.Error(w, "tenant not found", http.StatusNotFound)

        case errors.Is(err, job.ErrJobExists):
            http.Error(w, "job already exists", http.StatusConflict)

        case errors.Is(err, service.ErrTenantMismatch):
            http.Error(w, "tenant mismatch", http.StatusForbidden)

        case errors.Is(err, job.ErrInvalidJob),
            errors.Is(err, job.ErrEmptySteps),
            errors.Is(err, job.ErrEmptyStepID),
            errors.Is(err, job.ErrDuplicateStepID):
            http.Error(w, err.Error(), http.StatusBadRequest)

        default:
            http.Error(w, "internal server error", http.StatusInternalServerError)
        }

        return
    }

    // Return success.
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)

    json.NewEncoder(w).Encode(map[string]string{
        "message": "job submitted successfully",
        "job_id":  j.ID,
    })
}