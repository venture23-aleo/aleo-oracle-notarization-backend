package attestation

import (
	"bytes"
	"encoding/binary"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
)

// AttestationRequest is the request body for the attestation service.
type AttestationRequest struct {
	Url string `json:"url"`

	RequestMethod  string  `json:"requestMethod" validate:"required"`
	Selector       string  `json:"selector,omitempty" validate:"required"`
	ResponseFormat string  `json:"responseFormat" validate:"required"`
	HTMLResultType *string `json:"htmlResultType,omitempty" validate:"required"`

	RequestBody        *string `json:"requestBody,omitempty"`
	RequestContentType *string `json:"requestContentType,omitempty" validate:"required"`

	RequestHeaders map[string]string `json:"requestHeaders,omitempty"`

	EncodingOptions encoding.EncodingOptions `json:"encodingOptions" validate:"required"`

	DebugRequest bool `json:"debugRequest,omitempty"`
}

// AttestationResponse is the response body for the attestation service.
type AttestationResponse struct {
	ReportType string `json:"reportType"`

	AttestationRequest AttestationRequest `json:"attestationRequest"`

	AttestationReport string `json:"attestationReport"`

	AttestationTimestamp int64 `json:"timestamp"`

	ResponseBody string `json:"responseBody"`

	ResponseStatusCode int `json:"responseStatusCode"`

	AttestationData string `json:"attestationData"`

	OracleData OracleData `json:"oracleData"`
}

type AttestationRequestWithDebug struct {
	AttestationRequest

	DebugRequest bool `json:"debugRequest"`
}

type DebugAttestationResponse struct {
	ReportType string `json:"reportType"`

	AttestationRequest AttestationRequest `json:"attestationRequest"`

	AttestationTimestamp int64 `json:"timestamp"`

	ResponseBody string `json:"responseBody"`

	ResponseStatusCode int `json:"responseStatusCode"`

	AttestationData string `json:"attestationData"`
}

// Validate validates the attestation request.
func (ar *AttestationRequest) Validate(whitelistedDomains []string) *appErrors.AppError {

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
	if !isAcceptedDomain(ar.Url, whitelistedDomains) {
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
func isAcceptedDomain(endpoint string, whitelistedDomains []string) bool {
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
	for _, domainName := range whitelistedDomains {
		if domainName == parsedURL.Hostname() {
			return true
		}
	}
	return false
}

// wrapRawQuoteAsOpenEnclaveEvidence wraps the raw quote as Open Enclave evidence.
func wrapRawQuoteAsOpenEnclaveEvidence(rawQuoteBuffer []byte) []byte {

	const (
		OE_VERSION     = 1
		OE_VERSION_LEN = 4
		OE_TYPE        = 2
		OE_TYPE_LEN    = 4
		OE_QUOTE_LEN   = 8
	)

	// Create the Open Enclave version.
	oeVersion := make([]byte, OE_VERSION_LEN)
	binary.LittleEndian.PutUint32(oeVersion, OE_VERSION)

	// Create the Open Enclave type.
	oeType := make([]byte, OE_TYPE_LEN)
	binary.LittleEndian.PutUint32(oeType, OE_TYPE)

	// Create the quote length.
	quoteLength := make([]byte, OE_QUOTE_LEN)
	binary.LittleEndian.PutUint32(quoteLength, uint32(len(rawQuoteBuffer)))

	// Create the buffer.
	var buf bytes.Buffer

	// Write the Open Enclave version, type, and quote length to the buffer.
	buf.Write(oeVersion)
	buf.Write(oeType)
	buf.Write(quoteLength)
	buf.Write(rawQuoteBuffer)

	// Return the buffer as a byte slice.
	return buf.Bytes()
}

// quoteLock is the lock for the quote.
var quoteLock sync.Mutex

func GetQuoteLock() *sync.Mutex {
	return &quoteLock
}

// GenerateQuote generates a quote for the attestation service.
func GenerateQuote(inputData []byte) ([]byte, *appErrors.AppError) {

	// Lock the quote.
	common.GetQuoteLock().Lock()
	defer common.GetQuoteLock().Unlock()

	// Create the report data.
	reportData := make([]byte, 64)

	// Copy the input data to the report data.
	copy(reportData, inputData)

	// Write the report data to the user report data path.
	err := os.WriteFile(constants.GRAMINE_PATHS.USER_REPORT_DATA_PATH, reportData, 0644)

	if err != nil {
		logger.Error("Error while writting report data:", "error", err)
		return nil, appErrors.NewAppError(appErrors.ErrWrittingReportData)
	}

	// Read the quote from the quote path.
	quote, err := os.ReadFile(constants.GRAMINE_PATHS.QUOTE_PATH)

	if err != nil {
		logger.Error("Error while reading quote: ", "error", err)
		return nil, appErrors.NewAppError(appErrors.ErrReadingQuote)
	}

	// Wrap the raw quote as Open Enclave evidence.
	finalQuote := wrapRawQuoteAsOpenEnclaveEvidence(quote)

	// Return the final quote.
	return finalQuote, nil
}
