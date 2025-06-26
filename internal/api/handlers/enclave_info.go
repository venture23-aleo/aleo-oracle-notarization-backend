package handlers

import (
	"log"
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services"
)

// GetEnclaveInfo handles the request to get the enclave info.
func GetEnclaveInfo(aleoContext services.AleoPublicContext) http.HandlerFunc { 
	return func(w http.ResponseWriter, req *http.Request) {

		// Check if the request method is not GET.
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Generate a short request ID.
		requestId := utils.GenerateShortRequestID()

		// Get the enclave info.
		enclaveInfo, err := services.GetSgxInfo()

		// Check if the error is not nil.
		if err != nil {
			log.Print("Error getting enclave info:", err)
			utils.WriteJsonError(w, http.StatusInternalServerError, err.(appErrors.AppError), requestId)
			return
		}

		// Create the instance info.
		instanceInfo := services.InstanceInfo{
			ReportType:   "sgx",
			Info:         enclaveInfo,
			SignerPubKey: aleoContext.GetPublicKey(),
		}

		// Write the JSON success response.
		utils.WriteJsonSuccess(w, http.StatusOK, instanceInfo)
}
}
