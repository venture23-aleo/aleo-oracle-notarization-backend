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
}

type AppErrorWithRequestID struct {
	AppError
	RequestID string `json:"requestId,omitempty"`
}

// Error implements the error interface.
func (e AppError) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}

// NewAppError creates a new AppError with the given code and message
func NewAppError(code uint, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WithResponseStatusCode creates a new AppError with a specific response status code
func (e *AppError) WithResponseStatusCode(responseStatusCode int) *AppError {
	return &AppError{
		Code:               e.Code,
		Message:            e.Message,
		ResponseStatusCode: responseStatusCode,
		Details:            e.Details,
	}
}

// WithDetails creates a new AppError with a specific details
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:               e.Code,
		Message:            e.Message,
		ResponseStatusCode: e.ResponseStatusCode,
		Details:            details,
	}
}

// AppErrors are the errors for the application.
var (
	// =============================================================================
	// VALIDATION ERRORS (1000-1999)
	// =============================================================================
	ErrMissingURL                             = NewAppError(1001, "validation error: url is required")
	ErrMissingRequestMethod                   = NewAppError(1002, "validation error: requestMethod is required")
	ErrMissingResponseFormat                  = NewAppError(1003, "validation error: responseFormat is required")
	ErrMissingSelector                        = NewAppError(1004, "validation error: selector is required")
	ErrMissingEncodingOption                  = NewAppError(1005, "validation error: encodingOptions.value is required")
	ErrInvalidRequestMethod                   = NewAppError(1006, "validation error: requestMethod expected to be GET/POST")
	ErrMissingRequestBody                     = NewAppError(1007, "validation error: requestBody is required with POST requestMethod")
	ErrInvalidRequestBody                     = NewAppError(1008, "validation error: requestBody is not allowed with GET requestMethod")
	ErrInvalidResponseFormat                  = NewAppError(1009, "validation error: responseFormat expected to be html/json")
	ErrMissingHTMLResultType                  = NewAppError(1010, "validation error: htmlResultType is required with html responseFormat")
	ErrInvalidHTMLResultType                  = NewAppError(1011, "validation error: htmlResultType expected to be element/value")
	ErrInvalidEncodingOptionForHTMLResultType = NewAppError(1012, "validation error: expected encodingOptions.value to be string with htmlResultType element")
	ErrInvalidHTMLResultTypeForJSONResponse   = NewAppError(1013, "validation error: htmlResultType is not allowed with json responseFormat")
	ErrMissingEncodingValue                   = NewAppError(1014, "validation error: encodingOptions.value is required")
	ErrInvalidEncodingOption                  = NewAppError(1015, "validation error: invalid encoding option. expected: string/float/int")
	ErrTargetNotWhitelisted                   = NewAppError(1016, "validation error: attestation target is not whitelisted")
	ErrMissingEncodingPrecision               = NewAppError(1017, "validation error: encodingOptions.precision is required for float encoding")
	ErrInvalidEncodingPrecision               = NewAppError(1018, "validation error: invalid encodingOptions.precision")
	ErrMissingRequestContentType              = NewAppError(1019, "validation error: requestContentType is required with POST requestMethod")
	ErrInvalidRequestContentType              = NewAppError(1020, "validation error: requestContentType is not allowed with GET requestMethod")
	ErrInvalidSelector                        = NewAppError(1021, "validation error: selector expected to be weightedAvgPrice for price feed requests")
	ErrInvalidTargetURL                       = NewAppError(1022, "validation error: url should not include a scheme. Please remove https:// or http:// from your url")
	ErrMissingMaxParameter                    = NewAppError(1023, "validation error: missing max search parameter")
	ErrInvalidMaxParameter                    = NewAppError(1024, "validation error: invalid max search parameter")
	ErrInvalidMaxValue                        = NewAppError(1025, "validation error: expected max search parameter to be a number 2-2^127")
	ErrInvalidMaxValueFormat                  = NewAppError(1026, "validation error: invalid max value format")
	ErrInvalidAttestationData                 = NewAppError(1027, "validation error: attestation data is invalid")
	ErrInvalidURL                             = NewAppError(1028, "validation error: invalid url")
	ErrInvalidResponseFormatForPriceFeed      = NewAppError(1029, "validation error: responseFormat expected to be json for price feed requests")
	ErrInvalidEncodingOptionForPriceFeed      = NewAppError(1030, "validation error: invalid encoding option. expected: float for price feed requests")
	ErrInvalidRequestMethodForPriceFeed       = NewAppError(1031, "validation error: requestMethod expected to be GET for price feed requests")
	ErrInvalidSelectorForPriceFeed            = NewAppError(1032, "validation error: selector expected to be weightedAvgPrice for price feed requests")

	// =============================================================================
	// ENCLAVE ERRORS (2000-2999)
	// =============================================================================
	ErrReadingTargetInfo    = NewAppError(2001, "enclave error: failed to read the target info")
	ErrWrittingReportData   = NewAppError(2002, "enclave error: failed to write the report data")
	ErrGeneratingQuote      = NewAppError(2003, "enclave error: failed to generate the quote")
	ErrReadingQuote         = NewAppError(2004, "enclave error: failed to read the quote")
	ErrWrittingTargetInfo   = NewAppError(2005, "enclave error: failed to write the target info")
	ErrWrappingQuote        = NewAppError(2006, "enclave error: failed to wrap quote in openenclave format")
	ErrReadingReport        = NewAppError(2007, "enclave error: failed to read the report")
	ErrInvalidSGXReportSize = NewAppError(2008, "enclave error: invalid SGX report size")
	ErrParsingSGXReport     = NewAppError(2009, "enclave error: failed to parse SGX report")
	ErrInvalidSGXQuoteSize  = NewAppError(2010, "enclave error: invalid SGX quote size")

	// =============================================================================
	// ATTESTATION ERRORS (3000-3999)
	// =============================================================================
	ErrAleoContext                    = NewAppError(7001, "attestation error: failed to initialize Aleo context")
	ErrPreparingHashMessage           = NewAppError(3001, "attestation error: failed to prepare hash message for oracle data")
	ErrPreparingProofData             = NewAppError(3002, "attestation error: failed to prepare proof data")
	ErrFormattingProofData            = NewAppError(3003, "attestation error: failed to format proof data")
	ErrCreatingAttestationHash        = NewAppError(3004, "attestation error: failed to create attestation hash")
	ErrFormattingEncodedProofData     = NewAppError(3005, "attestation error: failed to format encoded proof data")
	ErrCreatingRequestHash            = NewAppError(3006, "attestation error: failed to create request hash")
	ErrCreatingTimestampedRequestHash = NewAppError(3007, "attestation error: failed to create timestamped request hash")
	ErrFormattingQuote                = NewAppError(3008, "attestation error: failed to format quote")
	ErrHashingReport                  = NewAppError(3009, "attestation error: failed to hash the oracle report")
	ErrGeneratingSignature            = NewAppError(3010, "attestation error: failed to generate signature")

	// =============================================================================
	// DATA EXTRACTION ERRORS (4000-4999)
	// =============================================================================
	ErrInvalidHTTPRequest      = NewAppError(4001, "data extraction error: invalid http request")
	ErrFetchingData            = NewAppError(4002, "data extraction error: failed to fetch the data from the provided endpoint")
	ErrInvalidStatusCode       = NewAppError(4003, "data extraction error: invalid status code returned from endpoint")
	ErrReadingHTMLContent      = NewAppError(4004, "data extraction error: failed to read HTML content from target url")
	ErrParsingHTMLContent      = NewAppError(4005, "data extraction error: failed to parse HTML content from target url")
	ErrReadingJSONResponse     = NewAppError(4006, "data extraction error: failed to read the json response from target url")
	ErrDecodingJSONResponse    = NewAppError(4007, "data extraction error: failed to decode JSON response from target url")
	ErrSelectorNotFound        = NewAppError(4008, "data extraction error: selector not found")
	ErrUnsupportedPriceFeedURL = NewAppError(4009, "data extraction error: unsupported price feed URL")
	ErrAttestationDataTooLarge = NewAppError(4010, "data extraction error: attestation data too large")
	ErrParsingFloatValue       = NewAppError(4011, "data extraction error: extracted value expected to be float but failed to parse as float")
	ErrParsingIntValue         = NewAppError(4012, "data extraction error: extracted value expected to be int but failed to parse as int")
	ErrEmptyAttestationData    = NewAppError(4013, "data extraction error: extracted attestation data is empty")

	// =============================================================================
	// ENCODING ERRORS (5000-5999)
	// =============================================================================
	ErrEncodingAttestationData = NewAppError(5001, "encoding error: failed to encode attestation data")
	ErrEncodingResponseFormat  = NewAppError(5002, "encoding error: failed to encode response format")
	ErrEncodingEncodingOptions = NewAppError(5003, "encoding error: failed to encode encoding options")
	ErrEncodingHeaders         = NewAppError(5004, "encoding error: failed to encode headers")
	ErrEncodingOptionalFields  = NewAppError(5005, "encoding error: failed to encode optional fields")
	ErrPreparingMetaHeader     = NewAppError(5006, "encoding error: error while preparing meta header")
	ErrWrittingAttestationData = NewAppError(5007, "encoding error: failed to write attestation data to buffer")
	ErrWrittingTimestamp       = NewAppError(5008, "encoding error: failed to write timestamp to buffer")
	ErrWrittingStatusCode      = NewAppError(5009, "encoding error: failed to write status code to buffer")
	ErrWrittingUrl             = NewAppError(5010, "encoding error: failed to write url to buffer")
	ErrWrittingSelector        = NewAppError(5011, "encoding error: failed to write selector to buffer")
	ErrWrittingResponseFormat  = NewAppError(5012, "encoding error: failed to write response format to buffer")
	ErrWrittingRequestMethod   = NewAppError(5013, "encoding error: failed to write request method to buffer")
	ErrWrittingEncodingOptions = NewAppError(5014, "encoding error: failed to write encoding options to buffer")
	ErrWrittingRequestHeaders  = NewAppError(5015, "encoding error: failed to write request headers to buffer")
	ErrWrittingOptionalFields  = NewAppError(5016, "encoding error: failed to write optional headers to buffer")
	ErrUserDataTooShort        = NewAppError(5017, "encoding error: userData too short for expected zeroing")
	ErrSliceToU128             = NewAppError(5018, "encoding error: failed to convert slice to u128")

	// =============================================================================
	// PRICE FEED ERRORS (6000-6999)
	// =============================================================================
	ErrTokenNotSupported         = NewAppError(6001, "price feed error: token not supported. Supported tokens: BTC, ETH, ALEO")
	ErrExchangeNotConfigured     = NewAppError(6002, "price feed error: exchange not configured")
	ErrSymbolNotConfigured       = NewAppError(6003, "price feed error: symbol not configured")
	ErrExchangeNotSupported      = NewAppError(6004, "price feed error: exchange not supported")
	ErrCreatingExchangeRequest   = NewAppError(6005, "price feed error: failed to create exchange request")
	ErrFetchingFromExchange      = NewAppError(6006, "price feed error: failed to fetch from exchange")
	ErrExchangeInvalidStatusCode = NewAppError(6007, "price feed error: invalid status code returned from exchange")
	ErrReadingExchangeResponse   = NewAppError(6008, "price feed error: failed to read exchange response")
	ErrDecodingExchangeResponse  = NewAppError(6009, "price feed error: failed to decode exchange response")
	ErrParsingExchangeResponse   = NewAppError(6010, "price feed error: failed to parse exchange response")
	ErrMissingDataInResponse     = NewAppError(6011, "price feed error: missing data in response")
	ErrParsingPrice              = NewAppError(6012, "price feed error: failed to parse price")
	ErrParsingVolume             = NewAppError(6013, "price feed error: failed to parse volume")
	ErrInsufficientExchangeData  = NewAppError(6014, "price feed error: insufficient data from exchanges")
	ErrEncodingPriceFeedData     = NewAppError(6015, "price feed error: failed to encode price feed data")
	ErrNoTradingPairsConfigured  = NewAppError(6016, "price feed error: no trading pairs configured for token")

	// =============================================================================
	// REQUEST/RESPONSE ERRORS (7000-7999)
	// =============================================================================
	ErrRequestBodyTooLarge = NewAppError(7001, "request error: payload exceeds the allowed size limit")
	ErrReadingRequestBody  = NewAppError(7002, "request error: failed to read the request body")
	ErrInvalidContentType  = NewAppError(7003, "request error: invalid content type, expected application/json")
	ErrDecodingRequestBody = NewAppError(7004, "request error: failed to decode request body, invalid request structure")

	// =============================================================================
	// INTERNAL ERRORS (8000-8999)
	// =============================================================================
	ErrInternal               = NewAppError(8001, "internal error: unexpected failure occurred")
	ErrGeneratingRandomNumber = NewAppError(8002, "internal error: failed to generate random number")
	ErrJSONEncoding           = NewAppError(8003, "internal error: failed to encode data to JSON")
)
