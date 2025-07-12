package errors

import (
	"fmt"
)

// AppError is the error for the application.
type AppError struct {
	Code               uint   `json:"errorCode"`
	Message            string `json:"errorMessage"`
	Details            string `json:"errorDetails,omitempty"`
	ResponseStatusCode int    `json:"responseStatusCode,omitempty"`
	RequestID          string `json:"requestId,omitempty"`
}

// Error implements the error interface.
func (e AppError) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}

// AppErrors are the errors for the application.
var (
	// =============================================================================
	// VALIDATION ERRORS (1000-1999)
	// =============================================================================
	ErrMissingURL               = AppError{1001, "validation error: url is required", "", 0, ""}
	ErrMissingRequestMethod     = AppError{1002, "validation error: requestMethod is required", "", 0, ""}
	ErrMissingResponseFormat    = AppError{1003, "validation error: responseFormat is required", "", 0, ""}
	ErrMissingSelector          = AppError{1004, "validation error: selector is required", "", 0, ""}
	ErrMissingEncodingOption    = AppError{1005, "validation error: encodingOptions.value is required", "", 0, ""}
	ErrInvalidRequestMethod     = AppError{1006, "validation error: requestMethod expected to be GET/POST", "", 0, ""}
	ErrMissingRequestBody       = AppError{1007, "validation error: requestBody is required with POST requestMethod", "", 0, ""}
	ErrInvalidResponseFormat    = AppError{1008, "validation error: responseFormat expected to be html/json", "", 0, ""}
	ErrMissingHTMLResultType    = AppError{1009, "validation error: htmlResultType is required with html responseFormat", "", 0, ""}
	ErrInvalidHTMLResultType    = AppError{1010, "validation error: htmlResultType expected to be element/value", "", 0, ""}
	ErrInvalidEncodingOption    = AppError{1011, "validation error: invalid encoding option", "", 0, ""}
	ErrUnacceptedDomain         = AppError{1012, "validation error: attestation target is not whitelisted", "", 0, ""}
	ErrInvalidRequestData       = AppError{1013, "validation error: failed to decode a request, invalid request structure", "", 0, ""}
	ErrInvalidEncodingPrecision = AppError{1014, "validation error: invalid encoding precision", "", 0, ""}

	// =============================================================================
	// ENCLAVE ERRORS (2000-2999)
	// =============================================================================
	ErrReadingTargetInfo  = AppError{2001, "enclave error: failed to read the target info", "", 0, ""}
	ErrWrittingReportData = AppError{2002, "enclave error: failed to write the report data", "", 0, ""}
	ErrGeneratingQuote    = AppError{2003, "enclave error: failed to generate the quote", "", 0, ""}
	ErrReadingQuote       = AppError{2004, "enclave error: failed to read the quote", "", 0, ""}
	ErrWrittingTargetInfo = AppError{2005, "enclave error: failed to write the target info", "", 0, ""}
	ErrWrappingQuote      = AppError{2006, "enclave error: failed to wrap quote in openenclave format", "", 0, ""}
	ErrReadingReport      = AppError{2007, "enclave error: failed to read the report", "", 0, ""}

	// =============================================================================
	// ATTESTATION ERRORS (3000-3999)
	// =============================================================================
	ErrPreparingOracleData            = AppError{3001, "attestation error: failed to prepare oracle data", "", 0, ""}
	ErrMessageHashing                 = AppError{3002, "attestation error: failed to prepare hash message for oracle data", "", 0, ""}
	ErrPreparingProofData             = AppError{3003, "attestation error: failed to prepare proof data", "", 0, ""}
	ErrFormattingProofData            = AppError{3004, "attestation error: failed to format proof data", "", 0, ""}
	ErrGeneratingAttestationHash      = AppError{3005, "attestation error: failed to generate to attestation hash", "", 0, ""}
	ErrPreparingEncodedProof          = AppError{3006, "attestation error: failed to prepare encoded request proof", "", 0, ""}
	ErrFormattingEncodedProofData     = AppError{3007, "attestation error: failed to format encoded proof data", "", 0, ""}
	ErrUserDataTooShort               = AppError{3008, "attestation error: userData too short for expected zeroing", "", 0, ""}
	ErrCreatingRequestHash            = AppError{3009, "attestation error: failed to create request hash", "", 0, ""}
	ErrCreatingTimestampedRequestHash = AppError{3010, "attestation error: failed to create timestamped request hash", "", 0, ""}
	ErrFormattingQuote                = AppError{3011, "attestation error: failed to format quote", "", 0, ""}
	ErrReportHashing                  = AppError{3012, "attestation error: failed to hash the oracle report", "", 0, ""}
	ErrGeneratingSignature            = AppError{3013, "attestation error: failed to generate signature", "", 0, ""}
	ErrDecodingQuote                  = AppError{3014, "attestation error: failed to decode quote", "", 0, ""}

	// =============================================================================
	// DATA EXTRACTION ERRORS (4000-4999)
	// =============================================================================
	ErrInvalidHTTPRequest      = AppError{4001, "data extraction error: invalid http request", "", 0, ""}
	ErrFetchingData            = AppError{4002, "data extraction error: failed to fetch the data from the provided endpoint", "", 0, ""}
	ErrReadingHTMLContent      = AppError{4003, "data extraction error: failed to read HTML content", "", 0, ""}
	ErrParsingHTMLContent      = AppError{4004, "data extraction error: failed to parse HTML content", "", 0, ""}
	ErrSelectorNotFound        = AppError{4005, "data extraction error: selector not found", "", 0, ""}
	ErrInvalidMap              = AppError{4006, "data extraction error: expected map but got something else", "", 0, ""}
	ErrKeyNotFound             = AppError{4007, "data extraction error: key not found", "", 0, ""}
	ErrJSONDecoding            = AppError{4008, "data extraction error: failed to decode the JSON response", "", 0, ""}
	ErrJSONEncoding            = AppError{4009, "data extraction error: failed to encode data to JSON", "", 0, ""}
	ErrInvalidSelectorPart     = AppError{4010, "data extraction error: invalid selector part", "", 0, ""}
	ErrExpectedArray           = AppError{4011, "data extraction error: expected array at key", "", 0, ""}
	ErrIndexOutOfBound         = AppError{4012, "data extraction error: index out of bounds", "", 0, ""}
	ErrUnsupportedPriceFeedURL = AppError{4013, "data extraction error: unsupported price feed URL", "", 0, ""}
	ErrAttestationDataTooLarge = AppError{4014, "data extraction error: attestation data too large", "", 0, ""}

	// =============================================================================
	// ENCODING ERRORS (5000-5999)
	// =============================================================================
	ErrEncodingAttestationData  = AppError{5001, "encoding error: failed to encode attestation data", "", 0, ""}
	ErrEncodingResponseFormat   = AppError{5002, "encoding error: failed to encode response format", "", 0, ""}
	ErrEncodingEncodingOptions  = AppError{5003, "encoding error: failed to encode encoding options", "", 0, ""}
	ErrEncodingHeaders          = AppError{5004, "encoding error: failed to encode headers", "", 0, ""}
	ErrEncodingOptionalFields   = AppError{5005, "encoding error: failed to encode optional fields", "", 0, ""}
	ErrPreparationCriticalError = AppError{5006, "encoding error: critical error while preparing data", "", 0, ""}
	ErrWrittingAttestationData  = AppError{5007, "encoding error: failed to write attestation data to buffer", "", 0, ""}
	ErrWrittingTimestamp        = AppError{5008, "encoding error: failed to write timestamp to buffer", "", 0, ""}
	ErrWrittingStatusCode       = AppError{5009, "encoding error: failed to write status code to buffer", "", 0, ""}
	ErrWrittingUrl              = AppError{5010, "encoding error: failed to write url to buffer", "", 0, ""}
	ErrWrittingSelector         = AppError{5011, "encoding error: failed to write selector to buffer", "", 0, ""}
	ErrWrittingResponseFormat   = AppError{5012, "encoding error: failed to write response format to buffer", "", 0, ""}
	ErrWrittingRequestMethod    = AppError{5013, "encoding error: failed to write request method to buffer", "", 0, ""}
	ErrWrittingEncodingOptions  = AppError{5014, "encoding error: failed to write encoding options to buffer", "", 0, ""}
	ErrWrittingRequestHeaders   = AppError{5015, "encoding error: failed to write request headers to buffer", "", 0, ""}
	ErrWrittingOptionalFields   = AppError{5016, "encoding error: failed to write optional headers to buffer", "", 0, ""}

	// =============================================================================
	// EXCHANGE ERRORS (6000-6999)
	// =============================================================================
	ErrMissingSymbol                 = AppError{6001, "exchange error: symbol parameter is required", "", 0, ""}
	ErrInvalidSymbol                 = AppError{6002, "exchange error: invalid symbol. Supported symbols: BTC, ETH, ALEO", "", 0, ""}
	ErrPriceFeedFailed               = AppError{6003, "exchange error: failed to get price feed data", "", 0, ""}
	ErrExchangeNotConfigured         = AppError{6004, "exchange error: exchange not configured", "", 0, ""}
	ErrSymbolNotSupportedByExchange  = AppError{6005, "exchange error: symbol not supported by exchange", "", 0, ""}
	ErrUnsupportedExchange           = AppError{6006, "exchange error: unsupported exchange", "", 0, ""}
	ErrExchangeFetchFailed           = AppError{6007, "exchange error: failed to fetch from exchange", "", 0, ""}
	ErrExchangeInvalidStatusCode     = AppError{6008, "exchange error: returned invalid status code", "", 0, ""}
	ErrExchangeResponseDecodeFailed  = AppError{6009, "exchange error: failed to decode exchange response", "", 0, ""}
	ErrExchangeResponseParseFailed   = AppError{6010, "exchange error: failed to parse exchange response", "", 0, ""}
	ErrInvalidPriceFormat            = AppError{6011, "exchange error: invalid price format", "", 0, ""}
	ErrInvalidVolumeFormat           = AppError{6012, "exchange error: invalid volume format", "", 0, ""}
	ErrInvalidExchangeResponseFormat = AppError{6013, "exchange error: invalid exchange response format", "", 0, ""}
	ErrInvalidDataFormat             = AppError{6014, "exchange error: invalid data format", "", 0, ""}
	ErrInvalidItemFormat             = AppError{6015, "exchange error: invalid item format", "", 0, ""}
	ErrNoDataInResponse              = AppError{6016, "exchange error: no data in response", "", 0, ""}
	ErrPriceParseFailed              = AppError{6017, "exchange error: failed to parse price", "", 0, ""}
	ErrVolumeParseFailed             = AppError{6018, "exchange error: failed to parse volume", "", 0, ""}
	ErrInsufficientExchangeData      = AppError{6019, "exchange error: insufficient data from exchanges", "", 0, ""}

	// =============================================================================
	// ALEO CONTEXT ERRORS (7000-7999)
	// =============================================================================
	ErrAleoContext = AppError{7001, "aleo context error: failed to initialize Aleo context", "", 0, ""}
)

// NewAppError creates a new AppError with the given code and message
func NewAppError(err AppError) *AppError {
	return &AppError{
		Code:               err.Code,
		Message:            err.Message,
		Details:            err.Details,
		ResponseStatusCode: err.ResponseStatusCode,
	}
}

// NewAppErrorWithResponseStatus creates a new AppError with a specific response status code
func NewAppErrorWithResponseStatus(err AppError, responseStatusCode int) *AppError {
	return &AppError{
		Code:               err.Code,
		Message:            err.Message,
		Details:            err.Details,
		ResponseStatusCode: responseStatusCode,
	}
}

// NewAppErrorWithResponseStatus creates a new AppError with a specific response status code
func NewAppErrorWithDetails(err AppError, details string) *AppError {
	return &AppError{
		Code:               err.Code,
		Message:            err.Message,
		Details:            details,
		ResponseStatusCode: err.ResponseStatusCode,
	}
}
