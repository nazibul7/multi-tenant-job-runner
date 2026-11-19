package main

import (
	"log"
	"net/http"

	"github.com/nazibul7/multi-tenant-job-runner/internal/handler"
	"github.com/nazibul7/multi-tenant-job-runner/internal/tenant"
)

func main() {
	tenantStore := tenant.NewStore()

	tenantHandler := handler.NewTenantHandler(tenantStore)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /tenants", tenantHandler.CreateTenant)

	log.Println("Server running on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
