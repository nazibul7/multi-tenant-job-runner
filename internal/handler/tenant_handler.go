package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nazibul7/multi-tenant-job-runner/internal/service"
	"github.com/nazibul7/multi-tenant-job-runner/internal/store"
	"github.com/nazibul7/multi-tenant-job-runner/internal/tenant"
)

type TenantHandler struct {
	store   *store.TenantStore
	service *service.TenantService
}

func NewTenantHandler(
	tenantStore *store.TenantStore,
	tenantService *service.TenantService,
) *TenantHandler {
	return &TenantHandler{
		store:   tenantStore,
		service: tenantService,
	}
}

func (h *TenantHandler) CreateTenant(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed",
			http.StatusMethodNotAllowed)
		return
	}

	var t tenant.Tenant

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&t); err != nil {
		http.Error(w, "invalid request body",
			http.StatusBadRequest)
		return
	}

	err := h.store.Create(t)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidTenant):
			http.Error(w, err.Error(),
				http.StatusBadRequest)

		case errors.Is(err, store.ErrTenantExists):
			http.Error(w, err.Error(),
				http.StatusConflict)

		default:
			http.Error(w, "internal server error",
				http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":      t.ID,
		"message": "tenant created",
	})
}

func (h *TenantHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := r.PathValue("id")

	usage, err := h.service.GetUsage(tenantID)
	if err != nil {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(usage)
}

func (h *TenantHandler) LiveView(w http.ResponseWriter, r *http.Request) {
	tenantID := r.PathValue("id")

	_, err := h.service.GetUsage(tenantID)
	if err != nil {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported",
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ticker.C:
			usage, err := h.service.GetUsage(tenantID)
			if err != nil {
				return
			}

			data, err := json.Marshal(usage)
			if err != nil {
				return
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return
			}

			flusher.Flush()
		}
	}
}
