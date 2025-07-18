package handlers

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	aleoContext "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/aleo_context"
	enclaveInfo "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/enclave_info"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

// GetEnclaveInfo handles the request to get the enclave info.
func GetEnclaveInfo(w http.ResponseWriter, req *http.Request) {
	// Get logger from context (request ID automatically included by middleware)
	reqLogger := logger.FromContext(req.Context())

	// Get SGX enclave info
	sgxInfo, err := enclaveInfo.GetSgxInfo()
	if err != nil {
		reqLogger.Error("Failed to get SGX enclave info", "error", err)
		utils.WriteJsonError(w, http.StatusInternalServerError, *err, "")
		return
	}

	// Get Aleo context
	aleoContext, ctxErr := aleoContext.GetAleoContext()
	if ctxErr != nil {
		reqLogger.Error("Failed to get Aleo context", "error", ctxErr)
		utils.WriteJsonError(w, http.StatusInternalServerError, *ctxErr, "")
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
