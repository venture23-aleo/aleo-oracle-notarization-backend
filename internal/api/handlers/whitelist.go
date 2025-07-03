package handlers

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// GetWhiteListedDomains handles the request to get the whitelisted domains.
func GetWhiteListedDomains(w http.ResponseWriter, req *http.Request) {
	// Write the JSON success response.
	utils.WriteJsonSuccess(w, http.StatusOK, configs.GetWhitelistedDomains())
}
