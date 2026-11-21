package main

import (
	"log"
	"net/http"

	"github.com/nazibul7/multi-tenant-job-runner/internal/handler"
	"github.com/nazibul7/multi-tenant-job-runner/internal/service"
	"github.com/nazibul7/multi-tenant-job-runner/internal/store"
)

func main() {
	tenantStore := store.NewTenantStore()
	jobStore := store.NewJobStore()

	tenantSerice := service.NewTenantService(tenantStore, jobStore)
	jobService := service.NewJobService(tenantStore, jobStore)

	tenantHandler := handler.NewTenantHandler(tenantStore, tenantSerice)
	jobHandler := handler.NewJobHandler(jobService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /tenants", tenantHandler.CreateTenant)

	mux.HandleFunc("POST /jobs", jobHandler.SubmitJob)
	mux.HandleFunc("GET /jobs/{id}", jobHandler.GetJob)

	mux.HandleFunc("GET /tenants/{id}/usage", tenantHandler.GetUsage)
	mux.HandleFunc("GET /tenants/{id}/live", tenantHandler.LiveView)

	log.Println("Server running on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
