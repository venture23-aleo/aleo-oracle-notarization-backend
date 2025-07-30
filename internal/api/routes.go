package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api/handler"
)

// RegisterRoutes registers the routes for the API.
func RegisterRoutes(mux *http.ServeMux) {

	// Register the info route.
	mux.HandleFunc("GET /info", handler.GetEnclaveInfo)

	// Register the whitelist route.
	mux.HandleFunc("GET /whitelist", handler.GetWhiteListedDomains)

	// Register the health check route.
	mux.HandleFunc("GET /health", handler.GetHealthCheck)

	// Register the notarization route.
	mux.HandleFunc("POST /notarize", handler.GenerateAttestationReport)

	// Register the random number route.
	mux.HandleFunc("GET /random", handler.GenerateAttestedRandom)

	// Register the root route.
	mux.HandleFunc("GET /", handler.GetRoot)
}

func RegisterMetricsRoute(mux *http.ServeMux) {
	mux.Handle("GET /metrics", promhttp.Handler())
}
