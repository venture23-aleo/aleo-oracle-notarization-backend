package api

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api/handlers"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/middlewares"
)

// RegisterRoutes registers the routes for the API.
func RegisterRoutes(mux *http.ServeMux) {

	attestationMiddleware := []middlewares.Middleware{
		middlewares.IPRateLimitMiddleware,
	}

	// Register the notarization route.
	mux.Handle("POST /notarize", middlewares.ChainFunc(handlers.GenerateAttestationReport, attestationMiddleware...))

	// Register the random number route.
	mux.Handle("GET /random", middlewares.ChainFunc(handlers.GenerateAttestedRandom, attestationMiddleware...))

	// Register the info route.
	mux.Handle("GET /info", middlewares.ChainFunc(handlers.GetEnclaveInfo))

	// Register the whitelist route.
	mux.Handle("GET /whitelist", middlewares.ChainFunc(handlers.GetWhiteListedDomains))

	mux.Handle("GET /", middlewares.ChainFunc(handlers.GetHealthCheck))
}
