package enclave_info

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/sgx"
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.InitLogger("DEBUG")

	// Run the tests
	// os.Exit(m.Run())
	m.Run()
}

// TestFormatSGXReport tests the creation of SGXEnclaveInfo from report data
func TestFormatSGXReport_Valid(t *testing.T) {

	testCases := []struct {
		name                    string
		mrenclave               []byte
		mrsigner                []byte
		isvprodid               []byte
		isvsvn                  []byte
		debugFlag               uint32
		expectedDebug           bool
		expectedSecurityVersion uint16
	}{
		{
			name:                    "SGX report with debug flag set to 0",
			mrenclave:               []byte("mrenclave12345678901234567890123"),
			mrsigner:                []byte("mrsigner123456789012345678901234"),
			isvprodid:               []byte{0x01, 0x00},
			isvsvn:                  []byte{0x01, 0x00},
			debugFlag:               0x0,
			expectedDebug:           false,
			expectedSecurityVersion: 0x1,
		},
		{
			name:                    "SGX report with debug flag set to 2 (debug enclave)",
			mrenclave:               bytes.Repeat([]byte{0x01}, 32),
			mrsigner:                bytes.Repeat([]byte{0x01}, 32),
			isvprodid:               []byte{0x00, 0x00},
			isvsvn:                  []byte{0x00, 0x00},
			debugFlag:               0x2,
			expectedDebug:           true,
			expectedSecurityVersion: 0x0,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Convert ProductID to 16-byte array for base64 encoding
			rawProdID := make([]byte, 16)
			copy(rawProdID, testCase.isvprodid)

			sgxReport := make([]byte, sgx.SGXReportSize)
			parsedSGXReport, err := sgx.ParseSGXReport(sgxReport)
			assert.Nil(t, err, "Failed to parse SGX report")
			assert.NotNil(t, parsedSGXReport, "Report data should not be nil")
			expectedUniqueID := base64.StdEncoding.EncodeToString(testCase.mrenclave)
			expectedSignerID := base64.StdEncoding.EncodeToString(testCase.mrsigner)
			expectedProductID := base64.StdEncoding.EncodeToString(rawProdID)
			copy(parsedSGXReport.Body.MREnclave[:], testCase.mrenclave)
			copy(parsedSGXReport.Body.MRSigner[:], testCase.mrsigner)
			copy(parsedSGXReport.Body.ISVProdID[:], testCase.isvprodid)
			copy(parsedSGXReport.Body.ISVSVN[:], testCase.isvsvn)
			parsedSGXReport.Body.Attributes.Flags = uint64(testCase.debugFlag)
			t.Logf("Parsed SGX Report: %+v", parsedSGXReport.Body.Attributes.Flags)
			formattedSGXInfo := formatSGXReport(parsedSGXReport)
			assert.Equal(t, testCase.expectedDebug, formattedSGXInfo.Debug)
			assert.Equal(t, testCase.expectedSecurityVersion, formattedSGXInfo.SecurityVersion)
			assert.NotEmpty(t, formattedSGXInfo.UniqueID)
			assert.NotEmpty(t, formattedSGXInfo.SignerID)
			assert.NotEmpty(t, formattedSGXInfo.ProductID)
			assert.Equal(t, expectedUniqueID, formattedSGXInfo.UniqueID)
			assert.Equal(t, expectedSignerID, formattedSGXInfo.SignerID)
			assert.Equal(t, expectedProductID, formattedSGXInfo.ProductID)
			assert.NotEmpty(t, formattedSGXInfo.Aleo)
		})
	}
}

// TestCreateAleoEncodedInfo tests the Aleo encoding functionality
func TestCreateAleoEncodedInfo(t *testing.T) {
	// // Create report data with zero values
	mrenclave := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	mrsigner := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	isvprodid := []byte{0x01, 0x00}
	isvsvn := []byte{0x01, 0x00}

	sgxReport := make([]byte, sgx.SGXReportSize)

	parsedSGXReport, err := sgx.ParseSGXReport(sgxReport)

	copy(parsedSGXReport.Body.MREnclave[:], mrenclave)
	copy(parsedSGXReport.Body.MRSigner[:], mrsigner)
	copy(parsedSGXReport.Body.ISVProdID[:], isvprodid)
	copy(parsedSGXReport.Body.ISVSVN[:], isvsvn)

	assert.Nil(t, err, "Failed to parse SGX report")
	assert.NotNil(t, parsedSGXReport, "Report data should not be nil")

	// Create Aleo encoded info
	aleoInfo := encodeForAleo(parsedSGXReport.Body)

	// Verify Aleo encoding
	assert.Equal(t, aleoInfo.UniqueID, "{ chunk_1: 1u128, chunk_2: 1u128 }")
	assert.Equal(t, aleoInfo.SignerID, "{ chunk_1: 1u128, chunk_2: 2u128 }")
	assert.Equal(t, aleoInfo.ProductID, "1u128")
}

// TestCreateAleoEncodedInfo_ZeroValues tests Aleo encoding with zero values
func TestCreateAleoEncodedInfo_ZeroValues(t *testing.T) {
	sgxReport := make([]byte, sgx.SGXReportSize)

	parsedSGXReport, err := sgx.ParseSGXReport(sgxReport)
	assert.Nil(t, err, "Failed to parse SGX report")
	assert.NotNil(t, parsedSGXReport, "Report data should not be nil")

	// Create Aleo encoded info
	aleoInfo := encodeForAleo(parsedSGXReport.Body)

	// Verify zero values are handled correctly
	assert.Equal(t, aleoInfo.UniqueID, "{ chunk_1: 0u128, chunk_2: 0u128 }")
	assert.Equal(t, aleoInfo.SignerID, "{ chunk_1: 0u128, chunk_2: 0u128 }")
	assert.Equal(t, "0u128", aleoInfo.ProductID)
}
