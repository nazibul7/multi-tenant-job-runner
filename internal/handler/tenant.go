package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nazibul7/multi-tenant-job-runner/internal/tenant"
)

type TenantHandler struct {
	store *tenant.Store
}

func NewTenantHandler(store *tenant.Store) *TenantHandler {
	return &TenantHandler{
		store: store,
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
		case errors.Is(err, tenant.ErrInvalidTenant):
			http.Error(w, err.Error(),
				http.StatusBadRequest)

		case errors.Is(err, tenant.ErrTenantExists):
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
