package enclave_info

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"sync"

	common "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/sgx"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

// AleoEncodedSGXInfo is the information about the SGX enclave for Aleo.
type AleoEncodedSGXInfo struct {
	UniqueID  string `json:"uniqueId"`  // Same as UniqueID but encoded for Aleo as 2 uint128
	SignerID  string `json:"signerId"`  // Same as SignerID but encoded for Aleo as 2 uint128
	ProductID string `json:"productId"` // Same as ProductID but encoded for Aleo as 1 uint128
}

// SGXInfo is the information about the SGX enclave.
type SGXEnclaveInfo struct {
	SecurityVersion uint16             `json:"securityVersion"` // Security version of the enclave. For SGX enclaves, this is the ISVSVN value.
	Debug           bool               `json:"debug"`           // If true, the report is for a debug enclave.
	UniqueID        string             `json:"uniqueId"`        // The unique ID for the enclave. For SGX enclaves, this is the MRENCLAVE value.
	SignerID        string             `json:"signerId"`        // The signer ID for the enclave. For SGX enclaves, this is the MRSIGNER value.
	ProductID       string             `json:"productId"`       // The Product ID for the enclave. For SGX enclaves, this is the ISVPRODID value.
	Aleo            AleoEncodedSGXInfo `json:"aleo"`            // Some of the SGX report values encoded for Aleo.
}

// EnclaveInfoResponse is the information about the enclave.
type EnclaveInfoResponse struct {
	ReportType   string         `json:"reportType"`   // The type of report.
	Info         SGXEnclaveInfo `json:"info"`         // The SGX Enclave info.
	SignerPubKey string         `json:"signerPubKey"` // The signer public key.
}

// formatSgxReport formats the SGX report
func formatSGXReport(sgxReport *sgx.SGXReport) (SGXEnclaveInfo, *appErrors.AppError) {

	// Convert ProductID to 16-byte array for base64 encoding
	rawProdID := make([]byte, 16)
	copy(rawProdID, sgxReport.Body.ISVProdID[:])

	// Create Aleo-encoded values
	aleoInfo, err := encodeForAleo(sgxReport.Body)
	if err != nil {
		logger.Error("Failed to encode SGX info for Aleo", "error", err)
		return SGXEnclaveInfo{}, err
	}

	// Create the SGX info
	formattedSGXInfo := SGXEnclaveInfo{
		UniqueID:        base64.StdEncoding.EncodeToString(sgxReport.Body.MREnclave[:]),
		SignerID:        base64.StdEncoding.EncodeToString(sgxReport.Body.MRSigner[:]),
		ProductID:       base64.StdEncoding.EncodeToString(rawProdID),
		SecurityVersion: binary.LittleEndian.Uint16(sgxReport.Body.ISVSVN[:]),
		Debug:           sgxReport.Body.Attributes.Flags&sgx.DebugFlagMask == sgx.DebugFlagMask,
		Aleo:            aleoInfo,
	}

	return formattedSGXInfo, nil
}

// encodeForAleo encodes the SGX info for Aleo
func encodeForAleo(reportBody sgx.ReportBody) (AleoEncodedSGXInfo, *appErrors.AppError) {

	mrEnclaveChunk1, err := common.SliceToU128(reportBody.MREnclave[0:16])
	if err != nil {
		logger.Error("Failed to convert MRENCLAVE to uint128", "error", err)
		return AleoEncodedSGXInfo{}, err
	}

	mrEnclaveChunk2, err := common.SliceToU128(reportBody.MREnclave[16:32])
	if err != nil {
		logger.Error("Failed to convert MRSIGNER to uint128", "error", err)
		return AleoEncodedSGXInfo{}, err
	}

	mrSignerChunk1, err := common.SliceToU128(reportBody.MRSigner[0:16])
	if err != nil {
		logger.Error("Failed to convert MRSIGNER to uint128", "error", err)
		return AleoEncodedSGXInfo{}, err
	}

	mrSignerChunk2, err := common.SliceToU128(reportBody.MRSigner[16:32])
	if err != nil {
		logger.Error("Failed to convert MRSIGNER to uint128", "error", err)
		return AleoEncodedSGXInfo{}, err
	}

	// Create Aleo info
	aleoInfo := AleoEncodedSGXInfo{
		UniqueID:  fmt.Sprintf("{ chunk_1: %su128, chunk_2: %su128 }", mrEnclaveChunk1, mrEnclaveChunk2),
		SignerID:  fmt.Sprintf("{ chunk_1: %su128, chunk_2: %su128 }", mrSignerChunk1, mrSignerChunk2),
		ProductID: fmt.Sprintf("%du128", binary.LittleEndian.Uint16(reportBody.ISVProdID[:])),
	}

	return aleoInfo, nil
}

// Global singleton instances with lazy initialization
var (
	sgxEnclaveInfoOnce sync.Once           // Once for lazy initialization.
	sgxEnclaveInfo     SGXEnclaveInfo      // SGX Enclave info.
	sgxEnclaveInfoErr  *appErrors.AppError // SGX Enclave info error.
)

// initializeSGXInfo initializes the SGX enclave info along with the Aleo encoded info
func initializeSGXEnclaveInfo() (SGXEnclaveInfo, *appErrors.AppError) {
	logger.Debug("Starting SGX info initialization process")

	// Generate SGX report
	rawSgxReport, err := sgx.GenerateSGXReport()
	if err != nil {
		return SGXEnclaveInfo{}, err
	}

	// Parse SGX report
	parsedSGXReport, err := sgx.ParseSGXReport(rawSgxReport)
	if err != nil {
		return SGXEnclaveInfo{}, err
	}

	// Format SGX info from parsed enclave info
	formattedSgxInfo, err := formatSGXReport(parsedSGXReport)
	if err != nil {
		return SGXEnclaveInfo{}, err
	}

	logger.Debug("SGX info initialized successfully")
	return formattedSgxInfo, nil
}

// GetSGXEnclaveInfo gets the SGX Enclave info for the instance.
// Uses singleton pattern with lazy initialization - reads enclave data once, reuses for all requests
func GetSGXEnclaveInfo() (SGXEnclaveInfo, *appErrors.AppError) {
	sgxEnclaveInfoOnce.Do(func() {
		sgxEnclaveInfo, sgxEnclaveInfoErr = initializeSGXEnclaveInfo()
	})
	return sgxEnclaveInfo, sgxEnclaveInfoErr
}
