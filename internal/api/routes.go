package api

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api/handlers"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services"
)

// RegisterRoutes registers the routes for the API.
func RegisterRoutes(mux *http.ServeMux, aleoCtx services.AleoPublicContext) {
	// Register the routes.

	// Register the notarization route.
	mux.HandleFunc("POST /notarize", handlers.GenerateAttestationReportHandler(aleoCtx))

	// Register the info route.
	mux.HandleFunc("GET /info", handlers.GetEnclaveInfo(aleoCtx))

	// Register the whitelist route.
	mux.HandleFunc("GET/whitelist", handlers.GetWhiteListedDomains)

	mux.HandleFunc("GET /", handlers.GetHealthCheck)
}
