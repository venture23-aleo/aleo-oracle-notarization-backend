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
	ErrMissingURL                 = AppError{1001, "url is required", "", 0, ""}
	ErrMissingRequestMethod       = AppError{1002, "requestMethod is required", "", 0, ""}
	ErrMissingSelector            = AppError{1003, "selector is required", "", 0, ""}
	ErrMissingEncodingOption      = AppError{1004, "encodingOptions.value is required", "", 0, ""}
	ErrInvalidRequestMethod       = AppError{1005, "requestMethod expected to be GET/POST", "", 0, ""}
	ErrMissingRequestBody         = AppError{1006, "requestBody is required with POST requestMethod", "", 0,""}
	ErrInvalidResponseFormat      = AppError{1007, "responseFormat expected to be html/json", "", 0,""}
	ErrMissingHTMLResultType      = AppError{1008, "htmlResultType is required with html responseFormat", "", 0,""}
	ErrInvalidHTMLResultType      = AppError{1009, "htmlResultType expected to be element/value", "", 0,""}
	ErrInvalidEncodingOption      = AppError{1010, "invalid encoding option", "", 0,""}
	ErrUnacceptedDomain           = AppError{1011, "attestation target is not whitelisted", "", 0,""}
	ErrInvalidRequestData         = AppError{1012, "failed to decode a request, invalid request structure", "", 0,""}
	ErrInvalidHTTPRequest         = AppError{1013, "invalid http request", "", 0,""}
	ErrPreparingOracleData        = AppError{1014, "failed to prepare oracle data", "", 0,""}
	ErrMessageHashing             = AppError{1015, "failed to prepare hash message", "", 0,""}
	ErrPreparingProofData         = AppError{1016, "failed to prepare proof data", "", 0,""}
	ErrFormattingProofData        = AppError{1017, "failed to format proof data", "", 0,""}
	ErrGeneratingAttestationHash  = AppError{1018, "failed to generate to attestation hash", "", 0,""}
	ErrPreparingEncodedProof      = AppError{1019, "failed to prepare encoded request proof", "", 0,""}
	ErrFormattingEncodedProofData = AppError{1020, "failed to format encoded proof data", "", 0,""}
	ErrUserDataTooShort           = AppError{1021, "userData too short for expected zeroing", "", 0,""}
	ErrCreatingRequestHash        = AppError{1022, "failed to create request  hash", "", 0,""}
	ErrCreatingTimestampedHash    = AppError{1023, "failed to create timestamped  hash", "", 0,""}
	ErrFormattingQuote            = AppError{1024, "failed to format quote", "", 0,""}
	ErrReportHashing              = AppError{1025, "failed to hash the oracle report", "", 0,""}
	ErrGeneratingSignature        = AppError{1026, "failed to generate signature", "", 0,""}
	ErrFetchingData               = AppError{1027, "failed to fetch the data from the provided endpoint", "", 0,""}
	ErrReadingHTMLContent         = AppError{1028, "failed to read HTML content", "", 0,""}
	ErrParsingHTMLContent         = AppError{1029, "failed to parse HTML content", "", 0,""}
	ErrSelectorNotFound           = AppError{1030, "selector not found", "", 0,""}
	ErrInvalidMap                 = AppError{1031, "expected map but got something else", "", 0,""}
	ErrKeyNotFound                = AppError{1032, "selector not found", "", 0,""}
	ErrJSONDecoding               = AppError{1033, "failed to decode the JSON response", "", 0,""}
	ErrJSONEncoding               = AppError{1034, "failed to encode data to JSON", "", 0,""}
	ErrReadingTargetInfo          = AppError{1035, "failed to read the target info", "", 0,""}
	ErrWrittingReportData         = AppError{1036, "failed to write the report data", "", 0,""}
	ErrGeneratingQuote            = AppError{1037, "failed to generate the quote", "", 0,""}
	ErrReadingQuote               = AppError{1038, "failed to read the quote", "", 0,""}
	ErrWrittingTargetInfo         = AppError{1039, "failed to write the target info", "", 0,""}
	ErrWrappingQuote              = AppError{1040, "failed to wrap quote in openenclave format", "", 0,""}
	ErrEncodingAttestationData    = AppError{1041, "failed to encode attestation data", "", 0,""}
	ErrEncodingResponseFormat     = AppError{1042, "failed to encode response format", "", 0,""}
	ErrEncodingEncodingOptions    = AppError{1043, "failed to encode encoding options", "", 0,""}
	ErrEncodingHeaders            = AppError{1044, "failed to encode headers", "", 0,""}
	ErrEncodingOptionalFields     = AppError{1045, "failed to encode optional fields", "", 0,""}
	ErrPreparationCriticalError   = AppError{1046, "critical error while preparing data", "", 0,""}
	ErrWrittingAttestationData    = AppError{1047, "failed to write attestation data to buffer", "", 0,""}
	ErrWrittingTimestamp          = AppError{1048, "failed to write timestamp to buffer", "", 0,""}
	ErrWrittingStatusCode         = AppError{1049, "failed to write status code to buffer", "", 0,""}
	ErrWrittingUrl                = AppError{1050, "failed to write url to buffer", "", 0,""}
	ErrWrittingSelector           = AppError{1051, "failed to write selector to buffer", "", 0,""}
	ErrWrittingResponseFormat     = AppError{1052, "failed to write response format to buffer", "", 0,""}
	ErrWrittingRequestMethod      = AppError{1053, "failed to write request method to buffer", "", 0,""}
	ErrWrittingEncodingOptions    = AppError{1054, "failed to write encoding options to buffer", "", 0,""}
	ErrWrittingRequestHeaders     = AppError{1055, "failed to write request headers to buffer", "", 0,""}
	ErrWrittingOptionalFields     = AppError{1056, "failed to write optinal headers to buffer", "", 0,""}
	ErrInvalidSelectorPart        = AppError{1057, "invalid selector part", "", 0,""}
	ErrExpectedArray              = AppError{1058, "expected array at key", "", 0,""}
	ErrIndexOutOfBound            = AppError{1059, "index out of bounds", "", 0,""}
	ErrMissingSymbol              = AppError{1060, "symbol parameter is required", "", 0,""}
	ErrInvalidSymbol              = AppError{1061, "invalid symbol. Supported symbols: BTC, ETH, ALEO", "", 0,""}
	ErrPriceFeedFailed            = AppError{1062, "failed to get price feed data", "", 0,""}
	ErrReadingReport              = AppError{1063, "failed to read the report", "", 0,""}
	ErrUnsupportedPriceFeedURL    = AppError{1064, "unsupported price feed URL", "", 0,""}
	// Exchange configuration errors
	ErrExchangeNotConfigured      = AppError{1065, "exchange not configured", "", 0,""}
	ErrSymbolNotSupportedByExchange = AppError{1066, "symbol not supported by exchange", "", 0,""}
	ErrUnsupportedExchange        = AppError{1067, "unsupported exchange", "", 0,""}
	// Exchange API errors
	ErrExchangeFetchFailed        = AppError{1068, "failed to fetch from exchange", "", 0,""}
	ErrExchangeInvalidStatusCode  = AppError{1069, "exchange returned invalid status code", "", 0,""}
	ErrExchangeResponseDecodeFailed = AppError{1070, "failed to decode exchange response", "", 0,""}
	ErrExchangeResponseParseFailed = AppError{1071, "failed to parse exchange response", "", 0,""}
	// Data format errors
	ErrInvalidPriceFormat         = AppError{1072, "invalid price format", "", 0,""}
	ErrInvalidVolumeFormat        = AppError{1073, "invalid volume format", "", 0,""}
	ErrInvalidExchangeResponseFormat = AppError{1074, "invalid exchange response format", "", 0,""}
	ErrInvalidDataFormat          = AppError{1075, "invalid data format", "", 0,""}
	ErrInvalidItemFormat          = AppError{1076, "invalid item format", "", 0,""}
	ErrNoDataInResponse           = AppError{1077, "no data in response", "", 0,""}
	// Data parsing errors
	ErrPriceParseFailed           = AppError{1078, "failed to parse price", "", 0,""}
	ErrVolumeParseFailed          = AppError{1079, "failed to parse volume", "", 0,""}
	// Price feed validation errors
	ErrInsufficientExchangeData   = AppError{1080, "insufficient data from exchanges", "", 0,""}
	// Encoding validation errors
	ErrInvalidEncodingPrecision   = AppError{1081, "invalid encoding precision", "", 0,""}
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