package handler

import (
	"net/http"

	httpUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/httputil"
)

// RootResponse is the response for the root endpoint.
type RootResponse struct {
	Service     string `json:"service"`
	Description string `json:"description"`
}

// GetRoot handles the request to get service information.
func GetRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Write the JSON success response with service information.
	httpUtil.WriteJsonSuccess(w, http.StatusOK, RootResponse{
		Service:     "Aleo Oracle Notarization Backend",
		Description: "SGX-based data attestation service for Aleo",
	})
}
