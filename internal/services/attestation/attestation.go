package attestation

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
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

type QuoteDecodeResponse struct {
	Data            string `json:"data"`
	SecurityVersion uint   `json:"securityVersion"`
	Debug           bool   `json:"debug"`
	UniqueID        string `json:"uniqueId"`
	SignerID        string `json:"signerId"`
	ProductID       string `json:"productId"`
	TCBStatus       uint   `json:"tcbStatus"`
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
	if endpoint == constants.PriceFeedBtcUrl || endpoint == constants.PriceFeedEthUrl || endpoint == constants.PriceFeedAleoUrl {
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

	// Create the Open Enclave version.
	oeVersion := make([]byte, 4)
	binary.LittleEndian.PutUint32(oeVersion, 1)

	// Create the Open Enclave type.
	oeType := make([]byte, 4)
	binary.LittleEndian.PutUint32(oeType, 2)

	// Create the quote length.
	quoteLength := make([]byte, 8)
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
	quoteLock.Lock()
	defer quoteLock.Unlock()

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

func DecodeQuote(verifierEndpoint string, quote []byte) (QuoteDecodeResponse, *appErrors.AppError) {

	// Create the client.
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = logger.Logger
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 3 * time.Second
	retryClient.RetryMax = 3

	// Create the request body as JSON
	requestBody := map[string]string{
		"quote": base64.StdEncoding.EncodeToString(quote),
	}

	// Marshal the request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("Error while marshaling request body: ", "error", err)
		return QuoteDecodeResponse{}, appErrors.NewAppError(appErrors.ErrDecodingQuote)
	}

	fullURL := fmt.Sprintf("%s/decode_quote", verifierEndpoint)
	logger.Debug("Sending quote decode request", "url", fullURL, "quote_length", len(quote))

	result, err := retryClient.Post(fullURL, "application/json", bytes.NewReader(jsonBody))

	if err != nil {
		logger.Error("Error while making HTTP request: ", "error", err, "url", fullURL)
		return QuoteDecodeResponse{}, appErrors.NewAppError(appErrors.ErrDecodingQuote)
	}

	defer result.Body.Close()

	// Log response status
	logger.Debug("Received response", "status_code", result.StatusCode, "url", fullURL)

	// Check if response status is not successful
	if result.StatusCode != http.StatusOK {
		// Read the error response body
		errorBody, _ := io.ReadAll(result.Body)
		logger.Error("HTTP request failed", "status_code", result.StatusCode, "error_body", string(errorBody), "url", fullURL)
		return QuoteDecodeResponse{}, appErrors.NewAppError(appErrors.ErrDecodingQuote)
	}

	// Read the response body
	responseBody, err := io.ReadAll(result.Body)
	if err != nil {
		logger.Error("Error while reading response body: ", "error", err, "url", fullURL)
		return QuoteDecodeResponse{}, appErrors.NewAppError(appErrors.ErrDecodingQuote)
	}

	// Check if response body is empty
	if len(responseBody) == 0 {
		logger.Error("Empty response body received", "url", fullURL)
		return QuoteDecodeResponse{}, appErrors.NewAppError(appErrors.ErrDecodingQuote)
	}

	quoteDecodeResponse := QuoteDecodeResponse{}

	err = json.Unmarshal(responseBody, &quoteDecodeResponse)
	if err != nil {
		logger.Error("Error while unmarshaling response: ", "error", err, "response_body", string(responseBody), "url", fullURL)
		return QuoteDecodeResponse{}, appErrors.NewAppError(appErrors.ErrDecodingQuote)
	}

	return quoteDecodeResponse, nil
}

func GetTCBStatus(verifierEndpoint string) (uint, *appErrors.AppError) {

	quote, err := GenerateQuote(nil)
	if err != nil {
		return 0, appErrors.NewAppError(appErrors.ErrGeneratingQuote)
	}

	quoteDecodeResponse, err := DecodeQuote(verifierEndpoint, quote)
	if err != nil {
		return 0, appErrors.NewAppError(appErrors.ErrDecodingQuote)
	}

	return quoteDecodeResponse.TCBStatus, nil
}
