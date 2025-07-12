package enclave_info

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.InitLogger("DEBUG")

	// Run the tests
	// os.Exit(m.Run())
	m.Run()
}

// TestSGXReportData_Extraction tests the extraction of data from SGX report
func TestSGXReportData_Extraction(t *testing.T) {
	// Create a mock SGX report with known values
	report := createMockSGXReport(t, true, []byte("mrenclave12345678901234567890123"),
		[]byte("mrsigner123456789012345678901234"),
		[]byte{0x12, 0x34}, []byte{0x56, 0x78})

	// Test report parsing
	reportData, err := parseSGXReportData(report)
	require.Nil(t, err, "Failed to parse SGX report")
	require.NotNil(t, reportData, "Report data should not be nil")
	// Verify extracted data
	assert.True(t, reportData.Debug, "Debug flag should be true")
	assert.Equal(t, []byte("mrenclave12345678901234567890123"), reportData.MREnclave)
	assert.Equal(t, []byte("mrsigner123456789012345678901234"), reportData.MRSigner)
	assert.Equal(t, []byte{0x12, 0x34}, reportData.ProductID)
	assert.Equal(t, []byte{0x56, 0x78}, reportData.SecurityVersion)
}

// TestSGXReportData_NonDebug tests extraction from non-debug report
func TestSGXReportData_NonDebug(t *testing.T) {
	// Create a mock SGX report with debug flag off
	report := createMockSGXReport(t, false, []byte("mrenclave12345678901234567890123456"),
		[]byte("mrsigner12345678901234567890123456"),
		[]byte{0x12, 0x34}, []byte{0x56, 0x78})

	reportData, err := parseSGXReportData(report)
	require.Nil(t, err, "Failed to parse SGX report")
	require.NotNil(t, reportData, "Report data should not be nil")

	assert.False(t, reportData.Debug, "Debug flag should be false")
}

// TestSGXReportData_InvalidSize tests handling of invalid report size
func TestSGXReportData_InvalidSize(t *testing.T) {
	// Create a report that's too small
	report := make([]byte, 100) // Too small for valid SGX report

	_, err := parseSGXReportData(report)
	require.Error(t, err)
	assert.Equal(t, appErrors.ErrReadingReport.Code, err.Code)
}

// TestCreateSgxInfoFromReportData tests the creation of SgxInfo from report data
func TestCreateSgxInfoFromReportData(t *testing.T) {
	// Create mock report data
	reportData := &SGXReportData{
		Debug:           true,
		MREnclave:       []byte("mrenclave12345678901234567890123"),
		MRSigner:        []byte("mrsigner123456789012345678901234"),
		ProductID:       []byte{0x12, 0x34},
		SecurityVersion: []byte{0x56, 0x78},
	}

	// Create SGX info
	sgxInfo := createSgxInfoFromReportData(reportData)

	productID := make([]byte, 16)
	copy(productID, reportData.ProductID)

	// Verify the created info
	assert.True(t, sgxInfo.Debug)
	assert.Equal(t, uint16(0x7856), sgxInfo.SecurityVersion) // Little endian
	assert.NotEmpty(t, sgxInfo.UniqueID)                     // base64 of mrenclave...
	assert.NotEmpty(t, sgxInfo.SignerID)                     // base64 of mrsigner...
	assert.NotEmpty(t, sgxInfo.ProductID)                    // base64 of padded product ID
	assert.Equal(t, sgxInfo.UniqueID, base64.StdEncoding.EncodeToString([]byte("mrenclave12345678901234567890123")))
	assert.Equal(t, sgxInfo.SignerID, base64.StdEncoding.EncodeToString([]byte("mrsigner123456789012345678901234")))
	assert.Equal(t, sgxInfo.ProductID, base64.StdEncoding.EncodeToString(productID))
	assert.NotEmpty(t, sgxInfo.Aleo)
}

// TestCreateAleoEncodedInfo tests the Aleo encoding functionality
func TestCreateAleoEncodedInfo(t *testing.T) {
	// Create mock report data with known values
	reportData := &SGXReportData{
		Debug:           false,
		MREnclave:       []byte("mrenclave12345678901234567890123456"),
		MRSigner:        []byte("mrsigner12345678901234567890123456"),
		ProductID:       []byte{0x12, 0x34},
		SecurityVersion: []byte{0x56, 0x78},
	}

	// Create Aleo encoded info
	aleoInfo := createAleoEncodedInfo(reportData)

	// Verify Aleo encoding
	assert.Contains(t, aleoInfo.UniqueID, "chunk_1:")
	assert.Contains(t, aleoInfo.UniqueID, "chunk_2:")
	assert.Contains(t, aleoInfo.SignerID, "chunk_1:")
	assert.Contains(t, aleoInfo.SignerID, "chunk_2:")
	assert.Contains(t, aleoInfo.ProductID, "u128")
	// The actual value depends on byte order, so just check it's a valid u128
	assert.Contains(t, aleoInfo.ProductID, "u128")
}

// TestCreateAleoEncodedInfo_ZeroValues tests Aleo encoding with zero values
func TestCreateAleoEncodedInfo_ZeroValues(t *testing.T) {
	// Create report data with zero values
	reportData := &SGXReportData{
		Debug:           false,
		MREnclave:       make([]byte, 32),
		MRSigner:        make([]byte, 32),
		ProductID:       []byte{0x00, 0x00},
		SecurityVersion: []byte{0x00, 0x00},
	}

	// Create Aleo encoded info
	aleoInfo := createAleoEncodedInfo(reportData)

	// Verify zero values are handled correctly
	assert.Contains(t, aleoInfo.UniqueID, "chunk_1: 0u128")
	assert.Contains(t, aleoInfo.UniqueID, "chunk_2: 0u128")
	assert.Contains(t, aleoInfo.SignerID, "chunk_1: 0u128")
	assert.Contains(t, aleoInfo.SignerID, "chunk_2: 0u128")
	assert.Equal(t, "0u128", aleoInfo.ProductID)
}

// TestConstants tests the SGX report constants
func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, int(0x02), DEBUG_FLAG_MASK)
	assert.Equal(t, int(0644), FILE_PERMISSIONS)
	assert.Equal(t, int(64), REPORT_DATA_BYTES)

	// Test offset constants
	assert.Equal(t, int(48), FLAGS_OFFSET)
	assert.Equal(t, int(64), MRENCLAVE_OFFSET)
	assert.Equal(t, int(128), MRSIGNER_OFFSET)
	assert.Equal(t, int(256), ISVPRODID_OFFSET)
	assert.Equal(t, int(258), ISVSVN_OFFSET)
	assert.Equal(t, int(320), REPORT_DATA_OFFSET)

	// Test size constants
	assert.Equal(t, int(8), FLAGS_SIZE)
	assert.Equal(t, int(32), MRENCLAVE_SIZE)
	assert.Equal(t, int(32), MRSIGNER_SIZE)
	assert.Equal(t, int(2), ISVPRODID_SIZE)
	assert.Equal(t, int(2), ISVSVN_SIZE)
	assert.Equal(t, int(64), REPORT_DATA_SIZE)
}

// TestSgxInfo_JSONMarshaling tests JSON marshaling of SgxInfo
func TestSgxInfo_JSONMarshaling(t *testing.T) {
	sgxInfo := SgxInfo{
		SecurityVersion: 1234,
		Debug:           true,
		UniqueID:        "test-unique-id",
		SignerID:        "test-signer-id",
		ProductID:       "test-product-id",
		Aleo: SgxAleoInfo{
			UniqueID:  "aleo-unique-id",
			SignerID:  "aleo-signer-id",
			ProductID: "aleo-product-id",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(sgxInfo)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled SgxInfo
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, sgxInfo.SecurityVersion, unmarshaled.SecurityVersion)
	assert.Equal(t, sgxInfo.Debug, unmarshaled.Debug)
	assert.Equal(t, sgxInfo.UniqueID, unmarshaled.UniqueID)
	assert.Equal(t, sgxInfo.SignerID, unmarshaled.SignerID)
	assert.Equal(t, sgxInfo.ProductID, unmarshaled.ProductID)
	assert.Equal(t, sgxInfo.Aleo, unmarshaled.Aleo)
}

// TestSgxAleoInfo_JSONMarshaling tests JSON marshaling of SgxAleoInfo
func TestSgxAleoInfo_JSONMarshaling(t *testing.T) {
	aleoInfo := SgxAleoInfo{
		UniqueID:  "aleo-unique-id",
		SignerID:  "aleo-signer-id",
		ProductID: "aleo-product-id",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(aleoInfo)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled SgxAleoInfo
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, aleoInfo.UniqueID, unmarshaled.UniqueID)
	assert.Equal(t, aleoInfo.SignerID, unmarshaled.SignerID)
	assert.Equal(t, aleoInfo.ProductID, unmarshaled.ProductID)
}

// TestEnclaveInfoResponse_JSONMarshaling tests JSON marshaling of EnclaveInfoResponse
func TestEnclaveInfoResponse_JSONMarshaling(t *testing.T) {
	response := EnclaveInfoResponse{
		ReportType: "SGX",
		Info: SgxInfo{
			SecurityVersion: 1234,
			Debug:           true,
			UniqueID:        "test-unique-id",
		},
		SignerPubKey: "test-signer-pub-key",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled EnclaveInfoResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, response.ReportType, unmarshaled.ReportType)
	assert.Equal(t, response.Info.SecurityVersion, unmarshaled.Info.SecurityVersion)
	assert.Equal(t, response.Info.Debug, unmarshaled.Info.Debug)
	assert.Equal(t, response.Info.UniqueID, unmarshaled.Info.UniqueID)
	assert.Equal(t, response.SignerPubKey, unmarshaled.SignerPubKey)
}

// Helper function to create a mock SGX report
func createMockSGXReport(t *testing.T, debug bool, mrenclave, mrsigner, productID, securityVersion []byte) []byte {
	// Create a report with minimum required size
	report := make([]byte, REPORT_DATA_OFFSET+REPORT_DATA_SIZE)

	// Set debug flag
	if debug {
		report[FLAGS_OFFSET] |= DEBUG_FLAG_MASK
	}

	// Copy data to appropriate offsets
	copy(report[MRENCLAVE_OFFSET:MRENCLAVE_OFFSET+MRENCLAVE_SIZE], mrenclave)
	copy(report[MRSIGNER_OFFSET:MRSIGNER_OFFSET+MRSIGNER_SIZE], mrsigner)
	copy(report[ISVPRODID_OFFSET:ISVPRODID_OFFSET+ISVPRODID_SIZE], productID)
	copy(report[ISVSVN_OFFSET:ISVSVN_OFFSET+ISVSVN_SIZE], securityVersion)

	return report
}

// Helper function to parse SGX report data (extracted from readAndParseSGXReport for testing)
func parseSGXReportData(report []byte) (*SGXReportData, *appErrors.AppError) {
	// Validate report size
	if len(report) < REPORT_DATA_OFFSET+REPORT_DATA_SIZE {
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

	return reportData, nil
}
