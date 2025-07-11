package enclave_info

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"sync"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

// SGX report is structured as follows:
/*
   | Field       | Offset | Size (bytes) |
   | ----------- | ------ | ------------ |
   | FLAGS       | 48     | 8            |
   | XFRM        | 56     | 8            |
   | MRENCLAVE   | 64     | 32           |
   | MRSIGNER    | 128    | 32           |
   | ISVPRODID   | 256    | 2            |
   | ISVSVN      | 258    | 2            |
   | REPORT DATA | 320    | 64           |
*/

// SGX Report structure constants
const (
	// Report field offsets and sizes
	FLAGS_OFFSET       = 48
	FLAGS_SIZE         = 8
	XFRM_OFFSET        = 56
	XFRM_SIZE          = 8
	MRENCLAVE_OFFSET   = 64
	MRENCLAVE_SIZE     = 32
	MRSIGNER_OFFSET    = 128
	MRSIGNER_SIZE      = 32
	ISVPRODID_OFFSET   = 256
	ISVPRODID_SIZE     = 2
	ISVSVN_OFFSET      = 258
	ISVSVN_SIZE        = 2
	REPORT_DATA_OFFSET = 320
	REPORT_DATA_SIZE   = 64

	// Debug flag mask
	DEBUG_FLAG_MASK = 0x02

	// File permissions
	FILE_PERMISSIONS = 0644

	// Report data size
	REPORT_DATA_BYTES = 64
)

// SgxAleoInfo is the information about the SGX enclave for Aleo.
type SgxAleoInfo struct {
	UniqueID  string `json:"uniqueId"`  // Same as UniqueID but encoded for Aleo as 2 uint128
	SignerID  string `json:"signerId"`  // Same as SignerID but encoded for Aleo as 2 uint128
	ProductID string `json:"productId"` // Same as ProductID but encoded for Aleo as 1 uint128
}

// SgxInfo is the information about the SGX enclave.
type SgxInfo struct {
	SecurityVersion uint16      `json:"securityVersion"` // Security version of the enclave. For SGX enclaves, this is the ISVSVN value.
	Debug           bool        `json:"debug"`           // If true, the report is for a debug enclave.
	UniqueID        string      `json:"uniqueId"`        // The unique ID for the enclave. For SGX enclaves, this is the MRENCLAVE value.
	SignerID        string      `json:"signerId"`        // The signer ID for the enclave. For SGX enclaves, this is the MRSIGNER value.
	ProductID       string      `json:"productId"`       // The Product ID for the enclave. For SGX enclaves, this is the ISVPRODID value.
	Aleo            SgxAleoInfo `json:"aleo"`            // Some of the SGX report values encoded for Aleo.
	TCBStatus       uint        `json:"tcbStatus"`       // The status of the enclave's TCB level.
}

// EnclaveInfoResponse is the information about the enclave.
type EnclaveInfoResponse struct {
	ReportType   string  `json:"reportType"`   // The type of report.
	Info         SgxInfo `json:"info"`         // The SGX info.
	SignerPubKey string  `json:"signerPubKey"` // The signer public key.
}

// SGXReportData contains the extracted data from the SGX report
type SGXReportData struct {
	Debug           bool
	MREnclave       []byte
	MRSigner        []byte
	ProductID       []byte
	SecurityVersion []byte
}

// Global singleton instances with lazy initialization
var (
	sgxInfoOnce sync.Once
	sgxInfo     SgxInfo
	sgxInfoErr  *appErrors.AppError
)

// GetSgxInfo gets the SGX info for the instance.
// Uses singleton pattern with lazy initialization - reads enclave data once, reuses for all requests
func GetSgxInfo() (SgxInfo, *appErrors.AppError) {
	sgxInfoOnce.Do(func() {
		sgxInfo, sgxInfoErr = loadSgxInfo()
	})
	return sgxInfo, sgxInfoErr
}

// loadSgxInfo performs the actual SGX info loading (called only once)
func loadSgxInfo() (SgxInfo, *appErrors.AppError) {
	logger.Debug("Starting SGX info loading process")

	// Step 2: Read and parse SGX report
	reportData, err := readAndParseSGXReport()
	if err != nil {
		return SgxInfo{}, err
	}

	// Step 3: Create SGX info from parsed data
	sgxInfo := createSgxInfoFromReportData(reportData)

	logger.Debug("SGX info loading completed successfully")
	return sgxInfo, nil
}

// getSgxReport gets the SGX report.
func getSgxReport() ([]byte, *appErrors.AppError) {

	attestation.GetQuoteLock().Lock()
	defer attestation.GetQuoteLock().Unlock()

	// Read the target info from target info path
	targetInfo, err := os.ReadFile(constants.GRAMINE_PATHS.MY_TARGET_INFO_PATH)
	if err != nil {
		logger.Error("Error reading target info: ", "error", err)
		return nil, appErrors.NewAppError(appErrors.ErrReadingTargetInfo)
	}

	// Write the target info to the target info path
	if err := os.WriteFile(constants.GRAMINE_PATHS.TARGET_INFO_PATH, targetInfo, FILE_PERMISSIONS); err != nil {
		logger.Error("Error writing target info: ", "error", err)
		return nil, appErrors.NewAppError(appErrors.ErrWrittingTargetInfo)
	}

	// Create and write report data
	reportData := make([]byte, REPORT_DATA_BYTES)
	if err := os.WriteFile(constants.GRAMINE_PATHS.USER_REPORT_DATA_PATH, reportData, FILE_PERMISSIONS); err != nil {
		logger.Error("Error writing report data: ", "error", err)
		return nil, appErrors.NewAppError(appErrors.ErrWrittingReportData)
	}

	// Read the report from the report path
	report, err := os.ReadFile(constants.GRAMINE_PATHS.REPORT_PATH)
	if err != nil {
		logger.Error("Error reading report: ", "error", err)
		return nil, appErrors.NewAppError(appErrors.ErrReadingReport)
	}

	logger.Debug("Attestation environment prepared successfully")
	return report, nil
}

// readAndParseSGXReport reads the SGX report and extracts the relevant data
func readAndParseSGXReport() (*SGXReportData, *appErrors.AppError) {

	// Get the report
	report, err := getSgxReport()
	if err != nil {
		return nil, err
	}

	// Validate report size
	if len(report) < REPORT_DATA_OFFSET+REPORT_DATA_SIZE {
		logger.Error("Report size too small: expected at least ", "expected", REPORT_DATA_OFFSET+REPORT_DATA_SIZE, "got", len(report))
		return nil, appErrors.NewAppError(appErrors.ErrReadingReport)
	}

	// Extract data from report
	reportData := &SGXReportData{
		Debug:           (report[FLAGS_OFFSET] & DEBUG_FLAG_MASK) > 0,
		MREnclave:       report[MRENCLAVE_OFFSET : MRENCLAVE_OFFSET+MRENCLAVE_SIZE],
		MRSigner:        report[MRSIGNER_OFFSET : MRSIGNER_OFFSET+MRSIGNER_SIZE],
		ProductID:       report[ISVPRODID_OFFSET : ISVPRODID_OFFSET+ISVPRODID_SIZE],
		SecurityVersion: report[ISVSVN_OFFSET : ISVSVN_OFFSET+ISVSVN_SIZE],
	}

	logger.Debug("Report parsed successfully - Debug: ", "debug", reportData.Debug, "securityVersion", binary.LittleEndian.Uint16(reportData.SecurityVersion))

	return reportData, nil
}

// createSgxInfoFromReportData creates the SgxInfo structure from parsed report data
func createSgxInfoFromReportData(reportData *SGXReportData) SgxInfo {

	// Convert ProductID to 16-byte array for base64 encoding
	rawProdID := make([]byte, 16)
	copy(rawProdID, reportData.ProductID)

	// Create Aleo-encoded values
	aleoInfo := createAleoEncodedInfo(reportData)

	// Create the SGX info
	sgxInfo := SgxInfo{
		UniqueID:        base64.StdEncoding.EncodeToString(reportData.MREnclave),
		SignerID:        base64.StdEncoding.EncodeToString(reportData.MRSigner),
		ProductID:       base64.StdEncoding.EncodeToString(rawProdID),
		SecurityVersion: binary.LittleEndian.Uint16(reportData.SecurityVersion),
		Debug:           reportData.Debug,
		Aleo:            aleoInfo,
		TCBStatus:       uint(configs.GetAppConfig().TCBStatus),
	}

	return sgxInfo
}

// createAleoEncodedInfo creates the Aleo-encoded information from SGX report data
func createAleoEncodedInfo(reportData *SGXReportData) SgxAleoInfo {

	// Convert MRENCLAVE to two uint128 chunks
	mrEnclaveChunk1 := new(big.Int).SetBytes(utils.ReverseBytes(reportData.MREnclave[:len(reportData.MREnclave)/2]))
	mrEnclaveChunk2 := new(big.Int).SetBytes(utils.ReverseBytes(reportData.MREnclave[len(reportData.MREnclave)/2:]))

	// Convert MRSIGNER to two uint128 chunks
	mrSignerChunk1 := new(big.Int).SetBytes(utils.ReverseBytes(reportData.MRSigner[:len(reportData.MRSigner)/2]))
	mrSignerChunk2 := new(big.Int).SetBytes(utils.ReverseBytes(reportData.MRSigner[len(reportData.MRSigner)/2:]))

	// Create Aleo info
	aleoInfo := SgxAleoInfo{
		UniqueID:  fmt.Sprintf("{ chunk_1: %su128, chunk_2: %su128 }", mrEnclaveChunk1, mrEnclaveChunk2),
		SignerID:  fmt.Sprintf("{ chunk_1: %su128, chunk_2: %su128 }", mrSignerChunk1, mrSignerChunk2),
		ProductID: fmt.Sprintf("%du128", binary.LittleEndian.Uint16(reportData.ProductID)),
	}

	return aleoInfo
}
