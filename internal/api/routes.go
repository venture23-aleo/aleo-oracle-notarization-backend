package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api/handlers"
)

// RegisterRoutes registers the routes for the API.
func RegisterRoutes(mux *http.ServeMux) {

	// Register the notarization route.
	mux.HandleFunc("POST /notarize", handlers.GenerateAttestationReport)

	// Register the random number route.
	mux.HandleFunc("GET /random", handlers.GenerateAttestedRandom)

	// Register the info route.
	mux.HandleFunc("GET /info", handlers.GetEnclaveInfo)

	// Register the whitelist route.
	mux.HandleFunc("GET /whitelist", handlers.GetWhiteListedDomains)

	// Register the Prometheus metrics endpoint.
	mux.Handle("GET /metrics", promhttp.Handler())

	// Register the health check route.
	mux.HandleFunc("GET /", handlers.GetHealthCheck)
}
