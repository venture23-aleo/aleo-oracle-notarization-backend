package handlers

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
)

// GetWhiteListedDomains handles the request to get the whitelisted domains.
func GetWhiteListedDomains(w http.ResponseWriter, req *http.Request) {

	// Check if the request method is not GET.
	if req.Method != http.MethodGet {
		// Write the status code.
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Write the JSON success response.
	utils.WriteJsonSuccess(w, http.StatusOK, configs.WHITELISTED_DOMAINS)
}
