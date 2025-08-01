package attestation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	// "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	// Initialize logger for tests
	// logger.InitLogger("DEBUG")

	// Run the tests
	m.Run()
}

func TestValidateAttestationRequestPayload_ValidCases(t *testing.T) {
	testCases := []struct {
		name               string
		attestationRequest AttestationRequest
		expectedError      *appErrors.AppError
	}{
		{
			name: "valid JSON request",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedError: nil,
		},
		{
			name: "valid HTML request with value type",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "body",
				HTMLResultType: &[]string{"value"}[0],
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: nil,
		},
		{
			name: "valid HTML request with element type",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "body",
				HTMLResultType: &[]string{"element"}[0],
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: nil,
		},
		{
			name: "valid POST request",
			attestationRequest: AttestationRequest{
				Url:                "google.com",
				RequestMethod:      "POST",
				ResponseFormat:     "json",
				Selector:           "body",
				RequestBody:        &[]string{`{"test": "data"}`}[0],
				RequestContentType: &[]string{"application/json"}[0],
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedError: nil,
		},
		{
			name: "valid price feed request",
			attestationRequest: AttestationRequest{
				Url:            "price_feed: btc",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "weightedAvgPrice",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedError: nil,
		},
		{
			name: "valid int encoding",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value: "int",
				},
			},
			expectedError: nil,
		},
		{
			name: "valid string encoding",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.attestationRequest.Validate()
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestValidateAttestationRequestPayload_InvalidCases(t *testing.T) {

	requestBody := `{"symbol": "BTC"}`
	requestContentType := "application/json"

	testCases := []struct {
		name               string
		attestationRequest AttestationRequest
		expectedError      *appErrors.AppError
	}{
		{
			name: "missing url",
			attestationRequest: AttestationRequest{
				RequestMethod:  "GET",
				ResponseFormat: "json",
				EncodingOptions: encoding.EncodingOptions{
					Precision: 8,
				},
			},
			expectedError: appErrors.ErrMissingURL,
		},
		{
			name: "missing request method",
			attestationRequest: AttestationRequest{
				Url: "www.google.com",
			},
			expectedError: appErrors.ErrMissingRequestMethod,
		},
		{
			name: "missing selector",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
			},
			expectedError: appErrors.ErrMissingSelector,
		},
		{
			name: "missing response format",
			attestationRequest: AttestationRequest{
				Url:           "www.google.com",
				RequestMethod: "GET",
				Selector:      "body",
			},
			expectedError: appErrors.ErrMissingResponseFormat,
		},
		{
			name: "missing encoding options",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
			},
			expectedError: appErrors.ErrMissingEncodingOption,
		},
		{
			name: "invalid response format",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "csv",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 8,
				},
			},
			expectedError: appErrors.ErrInvalidResponseFormat,
		},
		{
			name: "missing request body",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "POST",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 8,
				},
			},
			expectedError: appErrors.ErrMissingRequestBody,
		},
		{
			name: "missing encoding options",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
			},
			expectedError: appErrors.ErrMissingEncodingOption,
		},
		{
			name: "invalid request method",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "PUT",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "invalid",
					Precision: 8,
				},
			},
			expectedError: appErrors.ErrInvalidRequestMethod,
		},
		{
			name: "invalid encoding option",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "invalid",
					Precision: 8,
				},
			},
			expectedError: appErrors.ErrInvalidEncodingOption,
		},
		{
			name: "missing precision for float encoding option",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value: "float",
				},
			},
			expectedError: appErrors.ErrInvalidEncodingPrecision,
		},
		{
			name: "invalid precision for float encoding option",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 14,
				},
			},
			expectedError: appErrors.ErrInvalidEncodingPrecision,
		},
		{
			name: "invalid precision for float encoding option",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 14,
				},
			},
			expectedError: appErrors.ErrInvalidEncodingPrecision,
		},
		{
			name: "missing html result type",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedError: appErrors.ErrMissingHTMLResultType,
		},
		{
			name: "invalid html result type",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "body",
				HTMLResultType: &[]string{"invalid"}[0],
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedError: appErrors.ErrInvalidHTMLResultType,
		},
		{
			name: "invalid target url",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedError: appErrors.ErrInvalidTargetURL,
		},
		{
			name: "target not whitelisted",
			attestationRequest: AttestationRequest{
				Url:            "www.googles.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedError: appErrors.ErrTargetNotWhitelisted,
		},
		{
			name: "invalid request method for price feed",
			attestationRequest: AttestationRequest{
				Url:            "price_feed: btc",
				RequestMethod:  "POST",
				ResponseFormat: "json",
				Selector:       "weightedAvgPrice",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
				RequestBody:        &requestBody,
				RequestContentType: &requestContentType,
			},
			expectedError: appErrors.ErrInvalidRequestMethodForPriceFeed,
		},
		{
			name: "invalid selector for price feed",
			attestationRequest: AttestationRequest{
				Url:            "price_feed: btc",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedError: appErrors.ErrInvalidSelectorForPriceFeed,
		},
		{
			name: "invalid encoding option for price feed",
			attestationRequest: AttestationRequest{
				Url:            "price_feed: btc",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "weightedAvgPrice",
				EncodingOptions: encoding.EncodingOptions{
					Value: "int",
				},
			},
			expectedError: appErrors.ErrInvalidEncodingOptionForPriceFeed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.attestationRequest.Validate()
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}
func TestValidateAttestationRequestPayload_EdgeCases(t *testing.T) {
	testCases := []struct {
		name               string
		attestationRequest AttestationRequest
		expectedError      *appErrors.AppError
	}{
		{
			name: "float precision edge case - 0",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 0,
				},
			},
			expectedError: appErrors.ErrInvalidEncodingPrecision,
		},
		{
			name: "float precision edge case - 1",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 1,
				},
			},
			expectedError: nil,
		},
		{
			name: "float precision edge case - 12",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 12,
				},
			},
			expectedError: nil,
		},
		{
			name: "float precision edge case - 13",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 13,
				},
			},
			expectedError: appErrors.ErrInvalidEncodingPrecision,
		},
		{
			name: "int encoding with precision should fail",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "int",
					Precision: 6,
				},
			},
			expectedError: appErrors.ErrInvalidEncodingPrecision,
		},
		{
			name: "string encoding with precision should fail",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "string",
					Precision: 6,
				},
			},
			expectedError: appErrors.ErrInvalidEncodingPrecision,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.attestationRequest.Validate()
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestMaskUnacceptedHeaders_EdgeCases(t *testing.T) {
	testCases := []struct {
		name               string
		attestationRequest AttestationRequest
		expectedHeaders    map[string]string
	}{
		{
			name: "empty headers map",
			attestationRequest: AttestationRequest{
				RequestHeaders: map[string]string{},
			},
			expectedHeaders: map[string]string{},
		},
		{
			name: "nil headers map",
			attestationRequest: AttestationRequest{
				RequestHeaders: nil,
			},
			expectedHeaders: map[string]string{},
		},
		{
			name: "multiple unaccepted headers",
			attestationRequest: AttestationRequest{
				RequestHeaders: map[string]string{
					"X-Custom-Header":  "custom-value",
					"X-Test-Header":    "test-value",
					"X-Another-Header": "another-value",
				},
			},
			expectedHeaders: map[string]string{
				"X-Custom-Header":  "******",
				"X-Test-Header":    "******",
				"X-Another-Header": "******",
			},
		},
		{
			name: "mixed accepted and unaccepted headers",
			attestationRequest: AttestationRequest{
				RequestHeaders: map[string]string{
					"Content-Type":    "application/json",
					"X-Custom-Header": "custom-value",
					"Accept":          "application/json",
					"X-Test-Header":   "test-value",
				},
			},
			expectedHeaders: map[string]string{
				"Content-Type":    "application/json",
				"X-Custom-Header": "******",
				"Accept":          "application/json",
				"X-Test-Header":   "******",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.attestationRequest.MaskUnacceptedHeaders()
			assert.Equal(t, testCase.expectedHeaders, testCase.attestationRequest.RequestHeaders)
		})
	}
}

func TestAttestationResponse_Structure(t *testing.T) {
	// Test that AttestationResponse can be created and has expected fields
	response := AttestationResponse{
		ReportType: "test-report",
		AttestationRequest: AttestationRequest{
			Url:            "www.google.com",
			RequestMethod:  "GET",
			ResponseFormat: "json",
			Selector:       "body",
			EncodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 6,
			},
		},
		AttestationReport:    "test-report-data",
		AttestationTimestamp: 1234567890,
		ResponseBody:         "test-response-body",
		ResponseStatusCode:   200,
		AttestationData:      "test-attestation-data",
		OracleData: OracleData{
			Signature: "test-signature",
			UserData:  "test-user-data",
			Report:    "test-report",
			Address:   "test-address",
		},
	}

	assert.Equal(t, "test-report", response.ReportType)
	assert.Equal(t, "www.google.com", response.AttestationRequest.Url)
	assert.Equal(t, "test-report-data", response.AttestationReport)
	assert.Equal(t, int64(1234567890), response.AttestationTimestamp)
	assert.Equal(t, "test-response-body", response.ResponseBody)
	assert.Equal(t, 200, response.ResponseStatusCode)
	assert.Equal(t, "test-attestation-data", response.AttestationData)
	assert.Equal(t, "test-signature", response.OracleData.Signature)
}

func TestAttestationRequestWithDebug_Structure(t *testing.T) {
	// Test that AttestationRequestWithDebug can be created and has expected fields
	request := AttestationRequestWithDebug{
		AttestationRequest: AttestationRequest{
			Url:            "www.google.com",
			RequestMethod:  "GET",
			ResponseFormat: "json",
			Selector:       "body",
			EncodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 6,
			},
		},
		DebugRequest: true,
	}

	assert.Equal(t, "www.google.com", request.Url)
	assert.Equal(t, "GET", request.RequestMethod)
	assert.Equal(t, "json", request.ResponseFormat)
	assert.Equal(t, "body", request.Selector)
	assert.Equal(t, "float", request.EncodingOptions.Value)
	assert.Equal(t, uint(6), request.EncodingOptions.Precision)
	assert.True(t, request.DebugRequest)
}

func TestDebugAttestationResponse_Structure(t *testing.T) {
	// Test that DebugAttestationResponse can be created and has expected fields
	response := DebugAttestationResponse{
		ReportType: "debug-report",
		AttestationRequest: AttestationRequest{
			Url:            "www.google.com",
			RequestMethod:  "GET",
			ResponseFormat: "json",
			Selector:       "body",
			EncodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 6,
			},
		},
		AttestationTimestamp: 1234567890,
		ResponseBody:         "debug-response-body",
		ResponseStatusCode:   200,
		AttestationData:      "debug-attestation-data",
	}

	assert.Equal(t, "debug-report", response.ReportType)
	assert.Equal(t, "www.google.com", response.AttestationRequest.Url)
	assert.Equal(t, int64(1234567890), response.AttestationTimestamp)
	assert.Equal(t, "debug-response-body", response.ResponseBody)
	assert.Equal(t, 200, response.ResponseStatusCode)
	assert.Equal(t, "debug-attestation-data", response.AttestationData)
}

func TestOracleData_Structure(t *testing.T) {
	// Test that OracleData can be created and has expected fields
	oracleData := OracleData{
		Signature:              "test-signature",
		UserData:               "test-user-data",
		Report:                 "test-report",
		Address:                "test-address",
		EncodedPositions:       encoding.ProofPositionalInfo{},
		EncodedRequest:         "test-encoded-request",
		RequestHash:            "test-request-hash",
		TimestampedRequestHash: "test-timestamped-request-hash",
	}

	assert.Equal(t, "test-signature", oracleData.Signature)
	assert.Equal(t, "test-user-data", oracleData.UserData)
	assert.Equal(t, "test-report", oracleData.Report)
	assert.Equal(t, "test-address", oracleData.Address)
	assert.Equal(t, "test-encoded-request", oracleData.EncodedRequest)
	assert.Equal(t, "test-request-hash", oracleData.RequestHash)
	assert.Equal(t, "test-timestamped-request-hash", oracleData.TimestampedRequestHash)
}

// Benchmark tests for performance
func BenchmarkValidateAttestationRequest(b *testing.B) {
	request := AttestationRequest{
		Url:            "www.google.com",
		RequestMethod:  "GET",
		ResponseFormat: "json",
		Selector:       "body",
		EncodingOptions: encoding.EncodingOptions{
			Value:     "float",
			Precision: 6,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Validate()
	}
}

func BenchmarkMaskUnacceptedHeaders(b *testing.B) {
	request := AttestationRequest{
		RequestHeaders: map[string]string{
			"Content-Type":    "application/json",
			"X-Custom-Header": "custom-value",
			"Accept":          "application/json",
			"X-Test-Header":   "test-value",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.MaskUnacceptedHeaders()
	}
}

func TestIsAcceptedHeader(t *testing.T) {
	testCases := []struct {
		name     string
		header   string
		expected bool
	}{
		{name: "Content-Type", header: "Content-Type", expected: true},
		{name: "content-type", header: "content-type", expected: false},
		{name: "X-Custom-Header", header: "X-Custom-Header", expected: false},
		{name: "X-Custom-Header", header: "X-Test-Header", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, common.IsAcceptedHeader(testCase.header))
		})
	}
}

func TestMaskUnacceptedHeaders(t *testing.T) {

	testCases := []struct {
		name               string
		attestationRequest AttestationRequest
		expectedError      string
	}{
		{
			name: "no unaccepted headers",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedError: "",
		},
		{
			name: "unaccepted headers",
			attestationRequest: AttestationRequest{
				Url:            "www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
				RequestHeaders: map[string]string{
					"X-Custom-Header": "custom-value",
				},
			},
			expectedError: "******",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.attestationRequest.MaskUnacceptedHeaders()
			assert.Equal(t, testCase.expectedError, testCase.attestationRequest.RequestHeaders["X-Custom-Header"])
		})
	}
}

func createMockSGXQuote(reportData []byte, debug bool) []byte {
	mrenclave := []byte("mrenclave12345678901234567890123")
	mrsigner := []byte("mrsigner123456789012345678901234")
	productID := []byte{0x00, 0x01}
	securityVersion := []byte{0x00, 0x01}

	// Create a report with minimum required size
	report := make([]byte, 1450)

	// Set debug flag
	if debug {
		report[48] |= 0x02
	}

	// Copy data to appropriate offsets
	copy(report[64:64+32], mrenclave)
	copy(report[128:128+32], mrsigner)
	copy(report[256:256+2], productID)
	copy(report[258:258+2], securityVersion)
	copy(report[320:320+64], reportData)

	return report
}
