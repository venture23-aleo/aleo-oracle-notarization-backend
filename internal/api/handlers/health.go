package handlers

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

type HealthResponse struct {
	Message string `json:"messsage"`
}

func GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJsonSuccess(w,http.StatusOK, HealthResponse{Message: "Server running !!!"})
}
