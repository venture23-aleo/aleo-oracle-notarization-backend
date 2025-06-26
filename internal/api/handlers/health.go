package handlers

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// HealthResponse is the response for the health check.
type HealthResponse struct {
	// Message is the message of the health check.
	Message string `json:"messsage"`
}

// GetHealthCheck handles the request to get the health check.
func GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
	
	// Write the JSON success response.
	utils.WriteJsonSuccess(w, http.StatusOK, HealthResponse{Message: "Server running !!!"})
}
