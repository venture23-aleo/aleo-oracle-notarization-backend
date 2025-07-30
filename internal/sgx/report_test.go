package sgx

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

// TestParseSGXReport tests the parsing of SGX report
func TestParseSGXReport(t *testing.T) {

	mrenclave := []byte("mrenclave12345678901234567890123")
	mrsigner := []byte("mrsigner123456789012345678901234")
	isvprodid := []byte{0x01, 0x00}
	isvsvn := []byte{0x01, 0x00}

	report := CreateMockSGXReport(true, mrenclave, mrsigner, isvprodid, isvsvn)

	sgxReport, err := ParseSGXReport(report)
	require.Nil(t, err, "Failed to parse SGX report")
	require.NotNil(t, sgxReport, "Report data should not be nil")

	// Verify extracted data
	assert.True(t, sgxReport.Body.Attributes.Flags&DebugFlagMask == DebugFlagMask)
	assert.Equal(t, mrenclave, sgxReport.Body.MREnclave[:])
	assert.Equal(t, mrsigner, sgxReport.Body.MRSigner[:])
	assert.Equal(t, isvprodid, sgxReport.Body.ISVProdID[:])
	assert.Equal(t, isvsvn, sgxReport.Body.ISVSVN[:])
	assert.True(t, sgxReport.Body.Attributes.Flags&DebugFlagMask == DebugFlagMask)
}

func TestParseSGXReport_Valid(t *testing.T) {
	original := SGXReport{}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, &original)
	if err != nil {
		t.Fatalf("failed to serialize SGXReport: %v", err)
	}

	parsed, appErr := ParseSGXReport(buf.Bytes())
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}

	if !reflect.DeepEqual(*parsed, original) {
		t.Errorf("parsed report does not match original")
	}
}

// TestParseSGXReport_NonDebug tests parsing from non-debug report
func TestParseSGXReport_NonDebug(t *testing.T) {

	mrenclave := []byte("mrenclave12345678901234567890123")
	mrsigner := []byte("mrsigner123456789012345678901234")
	isvprodid := []byte{0x01, 0x00}
	isvsvn := []byte{0x01, 0x00}

	sgxReport := CreateMockSGXReport(false, mrenclave, mrsigner, isvprodid, isvsvn)

	reportData, err := ParseSGXReport(sgxReport)
	require.Nil(t, err, "Failed to parse SGX report")
	require.NotNil(t, reportData, "Report data should not be nil")

	assert.True(t, reportData.Body.Attributes.Flags&DebugFlagMask == DebugFlagMask)
}

// TestParseSGXReport_InvalidSize tests handling of invalid report size
func TestParseSGXReport_InvalidSize(t *testing.T) {
	// Create a report that's too small
	report := make([]byte, 100) // Too small for valid SGX report

	_, err := ParseSGXReport(report)
	assert.NotNil(t, err)
	assert.Equal(t, appErrors.ErrInvalidSGXReportSize.Code, err.Code)
}

// Helper function to create a mock SGX report
func CreateMockSGXReport(debug bool, mrenclave, mrsigner, productID, securityVersion []byte) []byte {

	// Create a report with minimum required size
	report := make([]byte, SGXReportSize)

	// Set debug flag
	if debug {
		report[48] |= 0x02
	}

	copy(report[48:48+8], []byte{0x02})
	copy(report[64:64+32], mrenclave)
	copy(report[128:128+32], mrsigner)
	copy(report[256:256+2], productID)
	copy(report[258:258+2], securityVersion)

	return report
}

func CreateMockReportWithDefaults() []byte {
	return CreateMockSGXReport(true, []byte("mrenclave12345678901234567890123"), []byte("mrsigner123456789012345678901234"), []byte{0x01, 0x00}, []byte{0x01, 0x00})
}
