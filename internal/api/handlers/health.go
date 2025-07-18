package handlers

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// HealthResponse is the response for the health check.
type HealthResponse struct {
	Status string `json:"status"`
}

// GetHealthCheck handles the request to get the health check.
func GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Write the JSON success response.
	utils.WriteJsonSuccess(w, http.StatusOK, HealthResponse{Status: "healthy"})
}
