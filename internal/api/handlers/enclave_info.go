package handlers

import (
	"log"
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services"
)

// GetEnclaveInfo handles the request to get the enclave info.
func GetEnclaveInfo(aleoContext services.AleoPublicContext) http.HandlerFunc { 
	return func(w http.ResponseWriter, req *http.Request) {
		// Generate a short request ID.
		requestId := utils.GenerateShortRequestID()

		// Get the enclave info.
		enclaveInfo, err := services.GetSgxInfo()

		// Check if the error is not nil.
		if err != nil {
			log.Print("Error getting enclave info:", err)
			utils.WriteJsonError(w, http.StatusInternalServerError, *err, requestId)
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
