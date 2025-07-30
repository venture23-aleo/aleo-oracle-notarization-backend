package handler

import (
	"net/http"

	configs "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/config"
	httpUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/httputil"
)

// GetWhiteListedDomains handles the request to get the whitelisted domains.
func GetWhiteListedDomains(w http.ResponseWriter, req *http.Request) {
	// Write the JSON success response.
	httpUtil.WriteJsonSuccess(w, http.StatusOK, configs.GetWhitelistedDomains())
}
