package api

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api/handlers"

	aleo "github.com/zkportal/aleo-utils-go"
)

// RegisterRoutes registers the routes for the API.
func RegisterRoutes(mux *http.ServeMux, s aleo.Session) {
	// Register the routes.

	// Register the notarization route.
	mux.HandleFunc("/notarize", handlers.GenerateAttestationReportHandler(s))

	// Register the info route.
	mux.HandleFunc("/info", handlers.GetInfo)

	// Register the whitelist route.
	mux.HandleFunc("/whitelist", handlers.GetWhiteListedDomains)


	mux.HandleFunc("/", handlers.GetHealthCheck)
}
