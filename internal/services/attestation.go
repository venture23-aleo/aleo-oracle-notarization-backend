package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	encoding "github.com/zkportal/aleo-oracle-encoding"
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

// Validate validates the attestation request.
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
	if ar.EncodingOptions.Value == "float" && (ar.EncodingOptions.Precision > encoding.ENCODING_OPTION_FLOAT_MAX_PRECISION) {
		return appErrors.NewAppError(appErrors.ErrInvalidEncodingPrecision)
	}

	// Check if the domain is accepted.
	if !utils.IsAcceptedDomain(ar.Url) {
		return appErrors.NewAppError(appErrors.ErrUnacceptedDomain)
	}

	return nil
}

// wrapRawQuoteAsOpenEnclaveEvidence wraps the raw quote as Open Enclave evidence.
func wrapRawQuoteAsOpenEnclaveEvidence(rawQuoteBuffer []byte) ([]byte, error) {

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
	return buf.Bytes(), nil
}

// dev/attestation/user_report_data -write
// dev/attestation/quote - read

// quoteLock is the lock for the quote.
var quoteLock sync.Mutex

// GenerateQuote generates a quote for the attestation service.
func GenerateQuote(inputData []byte) ([]byte, *appErrors.AppError) {

	// Lock the quote.
	quoteLock.Lock()
	defer quoteLock.Unlock()

	// Create the report data.
	reportData := make([]byte, 64)

	// Copy the input data to the report data.
	copy(reportData, inputData)

	fmt.Printf("64-byte report data: %x\n", reportData)

	// Write the report data to the user report data path.
	err := os.WriteFile(constants.GRAMINE_PATHS.USER_REPORT_DATA_PATH, reportData, 0644)

	if err != nil {
		log.Print("Error while writting report data:", err)
		return nil, appErrors.NewAppError(appErrors.ErrWrittingReportData)
	}

	// Read the quote from the quote path.
	quote, err := os.ReadFile(constants.GRAMINE_PATHS.QUOTE_PATH)

	if err != nil {
		log.Print("Generate Quote err: ", err)
		return nil, appErrors.NewAppError(appErrors.ErrReadingQuote)
	}

	// Wrap the raw quote as Open Enclave evidence.
	finalQuote, err := wrapRawQuoteAsOpenEnclaveEvidence(quote)

	if err != nil {
		return nil, appErrors.NewAppError(appErrors.ErrWrappingQuote)
	}

	// Return the final quote.
	return finalQuote, nil
}
