package attestation

import (
	"bytes"
	"encoding/binary"
	"net/url"
	"os"
	"strings"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
)

// AttestationRequest is the request body for the attestation service.
type AttestationRequest struct {
	Url string `json:"url"` // The URL to fetch data from.

	RequestMethod  string  `json:"requestMethod" validate:"required"`            // The request method.
	Selector       string  `json:"selector,omitempty" validate:"required"`       // The selector.
	ResponseFormat string  `json:"responseFormat" validate:"required"`           // The response format.
	HTMLResultType *string `json:"htmlResultType,omitempty" validate:"required"` // The HTML result type.

	RequestBody        *string `json:"requestBody,omitempty"`                            // The request body.
	RequestContentType *string `json:"requestContentType,omitempty" validate:"required"` // The request content type.

	RequestHeaders map[string]string `json:"requestHeaders,omitempty"` // The request headers.

	EncodingOptions encoding.EncodingOptions `json:"encodingOptions" validate:"required"` // The encoding options.

	DebugRequest bool `json:"debugRequest,omitempty"` // The debug request.
}

// AttestationResponse is the response body for the attestation service.
type AttestationResponse struct {
	ReportType string `json:"reportType"` // The report type.

	AttestationRequest AttestationRequest `json:"attestationRequest"` // The attestation request.

	AttestationReport string `json:"attestationReport"` // The attestation report.

	AttestationTimestamp int64 `json:"timestamp"` // The attestation timestamp.

	ResponseBody string `json:"responseBody"` // The response body.

	ResponseStatusCode int `json:"responseStatusCode"` // The response status code.

	AttestationData string `json:"attestationData"` // The attestation data.

	OracleData OracleData `json:"oracleData"` // The oracle data.
}

// AttestationRequestWithDebug is the attestation request with debug request.
type AttestationRequestWithDebug struct {
	AttestationRequest

	DebugRequest bool `json:"debugRequest"` // The debug request.
}

// DebugAttestationResponse is the debug attestation response.
type DebugAttestationResponse struct {
	ReportType string `json:"reportType"` // The report type.

	AttestationRequest AttestationRequest `json:"attestationRequest"` // The attestation request.

	AttestationTimestamp int64 `json:"timestamp"` // The attestation timestamp.

	ResponseBody string `json:"responseBody"` // The response body.

	ResponseStatusCode int `json:"responseStatusCode"` // The response status code.

	AttestationData string `json:"attestationData"` // The attestation data.
}

// Validate validates the attestation request and checks if target is whitelisted.
func (ar *AttestationRequest) Validate() *appErrors.AppError {

	// Check if the URL is empty.
	if ar.Url == "" {
		return appErrors.NewAppError(appErrors.ErrMissingURL)
	}

	// Check if the request method is empty.
	if ar.RequestMethod == "" {
		return appErrors.NewAppError(appErrors.ErrMissingRequestMethod)
	}

	// Check if the selector is empty.
	if ar.Selector == "" {
		return appErrors.NewAppError(appErrors.ErrMissingSelector)
	}

	if ar.ResponseFormat == "" {
		return appErrors.NewAppError(appErrors.ErrMissingResponseFormat)
	}

	// Check if the encoding option value is empty.
	if ar.EncodingOptions.Value == "" {
		return appErrors.NewAppError(appErrors.ErrMissingEncodingOption)
	}

	// Check if the request method is valid.
	if ar.RequestMethod != "GET" && ar.RequestMethod != "POST" {
		return appErrors.NewAppError(appErrors.ErrInvalidRequestMethod)
	}

	// Check if the request body is required for POST requests.
	if ar.RequestMethod == "POST" && ar.RequestBody == nil {
		return appErrors.NewAppError(appErrors.ErrMissingRequestBody)
	}

	// Check if the response format is valid.
	if ar.ResponseFormat != "html" && ar.ResponseFormat != "json" {
		return appErrors.NewAppError(appErrors.ErrInvalidResponseFormat)
	}

	// Check if the HTML result type is required for HTML response format.
	if ar.ResponseFormat == "html" && ar.HTMLResultType == nil {
		return appErrors.NewAppError(appErrors.ErrMissingHTMLResultType)
	}

	// Check if the HTML result type is valid.
	if ar.ResponseFormat == "html" && *ar.HTMLResultType != "value" && *ar.HTMLResultType != "element" {
		return appErrors.NewAppError(appErrors.ErrInvalidHTMLResultType)
	}

	// Check if the encoding option is valid.
	if ar.EncodingOptions.Value != "string" && ar.EncodingOptions.Value != "float" && ar.EncodingOptions.Value != "integer" {
		return appErrors.NewAppError(appErrors.ErrInvalidEncodingOption)
	}

	// Check if the encoding option precision is valid (only for float encoding).
	if ar.EncodingOptions.Value == "float" && (ar.EncodingOptions.Precision <= 0 || ar.EncodingOptions.Precision > encoding.ENCODING_OPTION_FLOAT_MAX_PRECISION) {
		return appErrors.NewAppError(appErrors.ErrInvalidEncodingPrecision)
	}

	// Check if the domain is accepted.
	if !isAcceptedDomain(ar.Url) {
		return appErrors.NewAppError(appErrors.ErrUnacceptedDomain)
	}

	return nil
}

// Masks unaccepted headers by replacing their values with "******"
func (ar *AttestationRequest) MaskUnacceptedHeaders() {
	finalHeaders := make(map[string]string)
	for headerName, headerValue := range ar.RequestHeaders {
		if !isAcceptedHeader(headerName) {
			finalHeaders[headerName] = "******"
		} else {
			finalHeaders[headerName] = headerValue
		}
	}
	ar.RequestHeaders = finalHeaders
}

// isAcceptedHeader checks if a header name is in the list of allowed headers.
func isAcceptedHeader(header string) bool {
	for _, h := range constants.ALLOWED_HEADERS {
		if strings.EqualFold(h, header) {
			return true
		}
	}
	return false
}

// isAcceptedDomain checks if a domain is in the list of whitelisted domains.
func isAcceptedDomain(endpoint string) bool {
	if endpoint == constants.PRICE_FEED_BTC_URL || endpoint == constants.PRICE_FEED_ETH_URL || endpoint == constants.PRICE_FEED_ALEO_URL {
		return true
	}

	var urlToParse string
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		urlToParse = endpoint
	} else {
		urlToParse = "https://" + endpoint
	}

	parsedURL, err := url.Parse(urlToParse)
	if err != nil {
		return false
	}
	for _, domainName := range configs.GetWhitelistedDomains() {
		if domainName == parsedURL.Hostname() {
			return true
		}
	}
	return false
}

// wrapRawQuoteAsOpenEnclaveEvidence wraps a raw SGX quote as Open Enclave evidence format.
//
// This function constructs an Open Enclave evidence buffer by prepending the required headers
// (version, type, and quote length) to the raw quote buffer. The resulting byte slice is formatted as follows:
//   - 4 bytes: Open Enclave version (little-endian uint32, currently 1)
//   - 4 bytes: Open Enclave type (little-endian uint32, currently 2)
//   - 8 bytes: Quote length (little-endian uint32, padded to 8 bytes)
//   - N bytes: Raw quote buffer
//
// This format is required by the verifier backend and the contract, which expects the quote to be wrapped as Open Enclave evidence.
//
// Parameters:
//   - rawQuoteBuffer ([]byte): The raw SGX quote to be wrapped.
//
// Returns:
//   - []byte: The Open Enclave evidence buffer containing the headers and the raw quote.
func wrapRawQuoteAsOpenEnclaveEvidence(rawQuoteBuffer []byte) []byte {

	const (
		OE_VERSION     = 1 // Open Enclave evidence version
		OE_VERSION_LEN = 4 // Length of version field in bytes
		OE_TYPE        = 2 // Open Enclave evidence type (2 = SGX ECDSA)
		OE_TYPE_LEN    = 4 // Length of type field in bytes
		OE_QUOTE_LEN   = 8 // Length of quote length field in bytes
	)

	// Create the Open Enclave version header (4 bytes, little-endian)
	oeVersion := make([]byte, OE_VERSION_LEN)
	binary.LittleEndian.PutUint32(oeVersion, OE_VERSION)

	// Create the Open Enclave type header (4 bytes, little-endian)
	oeType := make([]byte, OE_TYPE_LEN)
	binary.LittleEndian.PutUint32(oeType, OE_TYPE)

	// Create the quote length header (8 bytes, little-endian, only lower 4 bytes used)
	quoteLength := make([]byte, OE_QUOTE_LEN)
	binary.LittleEndian.PutUint32(quoteLength, uint32(len(rawQuoteBuffer)))

	// Assemble the Open Enclave evidence buffer
	var buf bytes.Buffer
	buf.Write(oeVersion)      // Write version header
	buf.Write(oeType)         // Write type header
	buf.Write(quoteLength)    // Write quote length header
	buf.Write(rawQuoteBuffer) // Write the raw quote

	// Return the complete Open Enclave evidence as a byte slice
	return buf.Bytes()
}

// GenerateQuote generates a quote for the attestation service.
//
// This function performs the following steps sequentially:
// 1. Acquires a lock to ensure thread-safe quote generation (as quote generation is not thread-safe).
// 2. Prepares a 64-byte report data buffer, copying the provided inputData into it (truncating or zero-padding as needed).
// 3. Writes the report data to the user report data path specified in the Gramine configuration.
// 4. Reads the raw SGX quote from the Gramine quote path.
// 5. Wraps the raw quote as Open Enclave evidence, as required by the contract and verifier backend.
// 6. Returns the wrapped quote as a byte slice, or an error if any step fails.
//
// Parameters:
//   - inputData ([]byte): The input data to be included in the report data (will be truncated or zero-padded to 64 bytes).
//
// Returns:
//   - []byte: The Open Enclave evidence buffer containing the wrapped quote.
//   - *appErrors.AppError: An application error if any step fails, otherwise nil.
func GenerateQuote(inputData []byte) ([]byte, *appErrors.AppError) {
	// Step 1: Acquire the quote lock for thread safety.
	common.GetQuoteLock().Lock()
	defer common.GetQuoteLock().Unlock()

	// Step 2: Prepare the 64-byte report data buffer.
	reportData := make([]byte, 64)
	copy(reportData, inputData) // Copy inputData (truncates or zero-pads as needed)

	// Step 3: Write the report data to the user report data path.
	err := os.WriteFile(constants.GRAMINE_PATHS.USER_REPORT_DATA_PATH, reportData, 0644)
	if err != nil {
		logger.Error("Error while writing report data:", "error", err)
		return nil, appErrors.NewAppError(appErrors.ErrWrittingReportData)
	}

	// Step 4: Read the raw quote from the Gramine quote path.
	quote, err := os.ReadFile(constants.GRAMINE_PATHS.QUOTE_PATH)
	if err != nil {
		logger.Error("Error while reading quote: ", "error", err)
		return nil, appErrors.NewAppError(appErrors.ErrReadingQuote)
	}

	// Step 5: Wrap the raw quote as Open Enclave evidence.
	finalQuote := wrapRawQuoteAsOpenEnclaveEvidence(quote)

	// Step 6: Return the final wrapped quote.
	return finalQuote, nil
}
