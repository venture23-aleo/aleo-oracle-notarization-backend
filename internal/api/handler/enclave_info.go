package handler

import (
	"net/http"

	aleoUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/aleoutil"
	httpUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/httputil"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	enclaveInfo "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/enclaveinfo"
)

// GetEnclaveInfo handles the request to get the enclave info.
func GetEnclaveInfo(w http.ResponseWriter, req *http.Request) {
	// Get logger from context (request ID automatically included by middleware)
	reqLogger := logger.FromContext(req.Context())

	// Get SGX enclave info
	sgxEnclaveInfo, err := enclaveInfo.GetSGXEnclaveInfo()
	if err != nil {
		reqLogger.Error("Failed to get SGX enclave info", "error", err)
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return
	}

	// Get Aleo context
	aleoContext, ctxErr := aleoUtil.GetAleoContext()
	if ctxErr != nil {
		reqLogger.Error("Failed to get Aleo context", "error", ctxErr)
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, ctxErr)
		return
	}

	// Create the instance info response
	enclaveInfoResponse := enclaveInfo.EnclaveInfoResponse{
		ReportType:   "sgx",
		Info:         sgxEnclaveInfo,
		SignerPubKey: aleoContext.GetPublicKey(),
	}

	httpUtil.WriteJsonSuccess(w, http.StatusOK, enclaveInfoResponse)
}
