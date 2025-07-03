package handlers

import (
	"log"
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
	
	enclaveInfo "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/enclave_info"
	aleoContext "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/aleo_context"

)

// GetEnclaveInfo handles the request to get the enclave info.
func GetEnclaveInfo(w http.ResponseWriter, req *http.Request) {
	// Generate a short request ID for tracing
	requestId := utils.GenerateShortRequestID()
	
	sgxInfo, err := enclaveInfo.GetSgxInfo()
	if err != nil {
		log.Printf("[%s] ERROR: Failed to get SGX enclave info: %v", requestId, err)
		utils.WriteJsonError(w, http.StatusInternalServerError, *err, requestId)
		return
	}

	aleoContext, ctxErr := aleoContext.GetAleoContext()
	if ctxErr != nil {
		log.Printf("[%s] ERROR: Failed to get Aleo context: %v", requestId, ctxErr)
		utils.WriteJsonError(w, http.StatusInternalServerError, *ctxErr, requestId)
		return
	}

	// Create the instance info response
	enclaveInfoResponse := enclaveInfo.EnclaveInfoResponse{
		ReportType:   "sgx",
		Info:         sgxInfo,
		SignerPubKey: aleoContext.GetPublicKey(),
	}
	
	utils.WriteJsonSuccess(w, http.StatusOK, enclaveInfoResponse)
}
