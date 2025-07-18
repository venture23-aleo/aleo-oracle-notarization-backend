package attestation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/enclave_info"
	// "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	// Initialize logger for tests
	// logger.InitLogger("DEBUG")

	// Run the tests
	m.Run()
}

func TestValidateAttestationRequestPayload(t *testing.T) {

	testCases := []struct {
		name               string
		attestationRequest AttestationRequest
		expectedPayload    *appErrors.AppError
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
			expectedPayload: appErrors.NewAppError(appErrors.ErrMissingURL),
		},
		{
			name: "missing request method",
			attestationRequest: AttestationRequest{
				Url: "https://www.google.com",
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrMissingRequestMethod),
		},
		{
			name: "missing selector",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrMissingSelector),
		},
		{
			name: "missing response format",
			attestationRequest: AttestationRequest{
				Url:           "https://www.google.com",
				RequestMethod: "GET",
				Selector:      "body",
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrMissingResponseFormat),
		},
		{
			name: "missing encoding options",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrMissingEncodingOption),
		},
		{
			name: "invalid response format",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "csv",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 8,
				},
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrInvalidResponseFormat),
		},
		{
			name: "missing request body",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "POST",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 8,
				},
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrMissingRequestBody),
		},
		{
			name: "missing encoding options",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrMissingEncodingOption),
		},
		{
			name: "invalid request method",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "PUT",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "invalid",
					Precision: 8,
				},
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrInvalidRequestMethod),
		},
		{
			name: "invalid encoding option",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "invalid",
					Precision: 8,
				},
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrInvalidEncodingOption),
		},
		{
			name: "missing precision for float encoding option",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value: "float",
				},
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrInvalidEncodingPrecision),
		},
		{
			name: "invalid precision for float encoding option",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 0,
				},
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrInvalidEncodingPrecision),
		},
		{
			name: "invalid precision for float encoding option",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 14,
				},
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrInvalidEncodingPrecision),
		},
		{
			name: "missing html result type",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrMissingHTMLResultType),
		},
		{
			name: "invalid html result type",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "body",
				HTMLResultType: &[]string{"invalid"}[0],
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrInvalidHTMLResultType),
		},
		{
			name: "domain not accepted",
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
			expectedPayload: appErrors.NewAppError(appErrors.ErrUnacceptedDomain),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.attestationRequest.Validate()
			if err != nil && err.Code != testCase.expectedPayload.Code {
				t.Fatalf("Expected error code %d, got %d", testCase.expectedPayload.Code, err.Code)
			}
		})
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
			assert.Equal(t, testCase.expected, isAcceptedHeader(testCase.header))
		})
	}
}

func TestMaskUnacceptedHeaders(t *testing.T) {

	testCases := []struct {
		name               string
		attestationRequest AttestationRequest
		expectedPayload    string
	}{
		{
			name: "no unaccepted headers",
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
			expectedPayload: "",
		},
		{
			name: "unaccepted headers",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
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
			expectedPayload: "******",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.attestationRequest.MaskUnacceptedHeaders()
			if testCase.attestationRequest.RequestHeaders["X-Custom-Header"] != testCase.expectedPayload {
				t.Fatalf("Expected %s, got %s", testCase.expectedPayload, testCase.attestationRequest.RequestHeaders["X-Custom-Header"])
			}
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
		report[enclave_info.FLAGS_OFFSET] |= enclave_info.DEBUG_FLAG_MASK
	}

	// Copy data to appropriate offsets
	copy(report[enclave_info.MRENCLAVE_OFFSET:enclave_info.MRENCLAVE_OFFSET+enclave_info.MRENCLAVE_SIZE], mrenclave)
	copy(report[enclave_info.MRSIGNER_OFFSET:enclave_info.MRSIGNER_OFFSET+enclave_info.MRSIGNER_SIZE], mrsigner)
	copy(report[enclave_info.ISVPRODID_OFFSET:enclave_info.ISVPRODID_OFFSET+enclave_info.ISVPRODID_SIZE], productID)
	copy(report[enclave_info.ISVSVN_OFFSET:enclave_info.ISVSVN_OFFSET+enclave_info.ISVSVN_SIZE], securityVersion)

	copy(report[enclave_info.REPORT_DATA_OFFSET:enclave_info.REPORT_DATA_OFFSET+enclave_info.REPORT_DATA_SIZE], reportData)

	return report
}
