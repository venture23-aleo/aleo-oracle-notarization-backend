package handler

import (
	"net/http"
	"time"

	httpUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/httputil"
)

// HealthResponse is the response for the health check.
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// GetHealthCheck handles the request to get the health check.
func GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Write the JSON success response.
	w.Header().Set("Cache-Control", "no-store")
	httpUtil.WriteJsonSuccess(w, http.StatusOK, HealthResponse{Status: "healthy", Timestamp: time.Now().Format(time.RFC3339)})
}
