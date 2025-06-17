package handlers

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
)

func GetWhiteListedDomains(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	utils.WriteJsonSuccess(w,http.StatusOK, configs.WHITELISTED_DOMAINS)
}