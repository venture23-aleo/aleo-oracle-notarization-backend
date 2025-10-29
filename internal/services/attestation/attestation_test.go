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
		ExtractedData:      "debug-attestation-data",
	}

	assert.Equal(t, "debug-report", response.ReportType)
	assert.Equal(t, "www.google.com", response.AttestationRequest.Url)
	assert.Equal(t, int64(1234567890), response.AttestationTimestamp)
	assert.Equal(t, "debug-response-body", response.ResponseBody)
	assert.Equal(t, 200, response.ResponseStatusCode)
	assert.Equal(t, "debug-attestation-data", response.ExtractedData)
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
		{name: "content-type", header: "content-type", expected: true},
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


func TestValidateHeaderValue(t *testing.T) {
	testCases := []struct {
		name     string
		field    string
		expected bool
	}{
		{name: "valid header", field: "application/json", expected: true},
		{name: "invalid header with carriage return and newline", field: "application/json\r\n", expected: false},
		{name: "invalid header with newline", field: "application/json\n", expected: false},
		{name: "invalid header with carriage return", field: "application/json\r", expected: false},
		{name: "valid header with custom value", field: "test-value", expected: true},
		{name: "valid header with numbers", field: "456", expected: true},
		{name: "valid header with dashes and underscores", field: "value_456", expected: true},
		{name: "valid header with spaces in value", field: "value with spaces", expected: true},
		{name: "valid header with tab in value", field: "value\twith\ttabs", expected: true},
		{name: "invalid header with control char in value", field: "value\x02", expected: false},
		{name: "valid header with colon in value", field: "value: value", expected: true},
		{name: "empty header value", field: "", expected: true},
		{name: "control character", field: "\x00", expected: false},
		{name: "percent-encoded CRLF", field: "%0d%0a", expected: false},
		{name: "percent-encoded CRLF", field: "%250d%250a", expected: false},
		{name: "percent-encoded CRLF", field: "%25250d%25250a", expected: false},
		{name: "percent-encoded CRLF", field: "%2525250d%2525250a", expected: false},
		{name: "percent-encoded CRLF", field: "%252525250d%252525250a", expected: false},
		{name: "percent-encoded CRLF", field: "%25252525250d%25252525250a", expected: false},
		{name: "percent-encoded CRLF", field: "%2525252525250d%2525252525250a", expected: false},
		{name: "percent-encoded CRLF", field: "%252525252525250d%252525252525250a", expected: false},
		{name: "unicode-style escapes", field: "%u000d", expected: false},
		{name: "unicode-style escapes", field: "%u000a", expected: false},
		{name: "unicode-style escapes", field: "%u000d%u000a", expected: false},
		{name: "unicode-style escapes", field: "%u000d%u000a", expected: false},
		{name: "unicode-style escapes", field: "%25u000d", expected: false},
		{name: "unicode-style escapes", field: "%25u000a", expected: false},
		{name: "unicode-style escapes", field: "%25u000d%25u000a", expected: false},
		{name: "unicode-style escapes", field: "%25u000d%25u000a", expected: false},
		{name: "unicode-style escapes", field: "%2525u000d", expected: false},
		{name: "unicode-style escapes", field: "%2525u000a", expected: false},
		{name: "unicode-style escapes", field: "%2525u000d%2525u000a", expected: false},
		{name: "unicode-style escapes", field: "%2525u000d%2525u000a", expected: false},
		{name: "unicode-style escapes", field: "%252525u000d", expected: false},
		{name: "unicode-style escapes", field: "%252525u000a", expected: false},
		{name: "unicode-style escapes", field: "%252525u000d%252525u000a", expected: false},
		{name: "unicode-style escapes", field: "\u000d", expected: false},
		{name: "unicode-style escapes", field: "\u000a", expected: false},
		{name: "unicode-style escapes", field: "\u000d\u000a", expected: false},
		{name: "unicode-style escapes", field: "\u000d\u000a", expected: false},
		{name: "unicode-style escapes", field: "%25\u000d", expected: false},
		{name: "unicode-style escapes", field: "%25\u000a", expected: false},
		{name: "unicode-style escapes", field: "%25\u000d%25\u000a", expected: false},
		{name: "hex style escapes", field: "\x00", expected: false},
		{name: "hex style escapes", field: "\x01", expected: false},
		{name: "hex style escapes", field: "\x02", expected: false},
		{name: "hex style escapes", field: "\x03", expected: false},
		{name: "hex style escapes", field: "\x04", expected: false},
		{name: "hex style escapes", field: "\x05", expected: false},
		{name: "hex style escapes", field: "\x06", expected: false},
		{name: "hex style escapes", field: "\x07", expected: false},
		{name: "hex style escapes", field: "\x08", expected: false},
		{name: "hex style escapes", field: "\x09", expected: true},
		{name: "hex style escapes", field: "\x0a", expected: false},
		{name: "hex style escapes", field: "\x0b", expected: false},
		{name: "hex style escapes", field: "\x0c", expected: false},
		{name: "hex style escapes", field: "\x0d", expected: false},
		{name: "hex style escapes", field: "\x0e", expected: false},
		{name: "hex style escapes", field: "\x0f", expected: false},
		{name: "hex style escapes", field: "\x10", expected: false},
		{name: "hex style escapes", field: "\x11", expected: false},
		{name: "hex style escapes", field: "\x12", expected: false},
		{name: "hex style escapes", field: "\x13", expected: false},
		{name: "hex style escapes", field: "\x14", expected: false},
		{name: "hex style escapes", field: "\x15", expected: false},
		{name: "hex style escapes", field: "\x16", expected: false},
		{name: "hex style escapes", field: "\x17", expected: false},
		{name: "hex style escapes", field: "\x18", expected: false},
		{name: "hex style escapes", field: "\x19", expected: false},
		{name: "hex style escapes", field: "\x1a", expected: false},
		{name: "hex style escapes", field: "\x1b", expected: false},
		{name: "hex style escapes", field: "\x1c", expected: false},
		{name: "hex style escapes", field: "\x1d", expected: false},
		{name: "hex style escapes", field: "\x1e", expected: false},
		{name: "hex style escapes", field: "\x1f", expected: false},
		{name: "encoded slash rn", field: "%255Cn", expected: false},
		{name: "encoded slash rn", field: "%255C%255Cn", expected: false},
		{name: "encoded slash rn", field: "%255C%255Cr", expected: false},
		{name: "encoded slash rn", field: "%255C%255Cr%255Cn", expected: false},
		{name: "encoded slash rn", field: "%255C%255Cr%255Cn%255C%255Cr", expected: false},
		{name: "encoded slash rn", field: "%255C%255Cr%255Cn%255C%255Cr%255C%255Cn", expected: false},
		{name: "encoded slash rn", field: "%255C%255Cr%255Cn%255C%255Cr%255C%255Cn%255C%255Cr", expected: false},
		{name: "encoded slash rn", field: "%255C%255Cr%255Cn%255C%255Cr%255C%255Cn%255C%255Cr%255C%255Cn", expected: false},
		{name: "unicode style escapes", field: "u000d", expected: true},
		{name: "unicode style escapes", field: "u000a", expected: true},
		{name: "unicode style escapes", field: "u000d%u000a", expected: false},
		{name: "unicode style escapes", field: "u000d%u000a", expected: false},
		{name: "percent as a text", field: "50% completed", expected: true},
		{name: "normal header value", field: "normal header value", expected: true},
		{name: "invalid utf-8", field: "\xff\x61\x62", expected: false},
		{name: "html char ref cr", field: "&#13;", expected: false},
		{name: "html char ref cr", field: "&#x0d;", expected: false},
		{name: "html char ref cr", field: "&#x0D;", expected: false},
		{name: "html char ref lf", field: "&#10;", expected: false},
		{name: "html char ref lf", field: "&#x0a;", expected: false},
		{name: "html char ref lf", field: "&#x0A;", expected: false},
		{name: "multiple percent", field: "%3%250e", expected: true},
	}


	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, isValidHeaderValue(testCase.field))
		})
	}
}


func TestValidateHeaderKey(t *testing.T) {
	testCases := []struct {
		name     string
		field    string
		expected bool
	}{
		{name: "valid header key", field: "Content-Type", expected: true},
		{name: "invalid header key", field: "Content-Type\r\n", expected: false},
		{name: "invalid header key", field: "Content-Type\n", expected: false},
		{name: "invalid header key", field: "Content-Type\r", expected: false},
		{name: "invalid header key", field: "", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, isValidHeaderKey(testCase.field))
		})
	}
}