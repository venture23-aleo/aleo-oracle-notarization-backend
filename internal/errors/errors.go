package errors

import "fmt"

// AppError is the error for the application.
type AppError struct {
	Code    int
	Message string
}

// Error returns the error message for the AppError.
func (e AppError) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}

// AppErrors are the errors for the application.
var (
	ErrMissingURL                 = AppError{1001, "url is required"}
	ErrMissingRequestMethod       = AppError{1002, "requestMethod is required"}
	ErrMissingSelector            = AppError{1003, "selector is required"}
	ErrMissingEncodingOption      = AppError{1004, "encodingOptions.value is required"}
	ErrInvalidRequestMethod       = AppError{1005, "requestMethod expected to be GET/POST"}
	ErrMissingRequestBody         = AppError{1006, "requestBody is required with POST requestMethod"}
	ErrInvalidResponseFormat      = AppError{1007, "responseFormat expected to be html/json"}
	ErrMissingHTMLResultType      = AppError{1008, "htmlResultType is required with html responseFormat"}
	ErrInvalidHTMLResultType      = AppError{1009, "htmlResultType expected to be element/value"}
	ErrInvalidEncodingOption      = AppError{1010, "invalid encoding option"}
	ErrUnacceptedDomain           = AppError{1011, "attestation target is not whitelisted"}
	ErrInvalidRequestData         = AppError{1012, "failed to decode a request, invalid request structure"}
	ErrInvalidHTTPRequest         = AppError{1013, "invalid http request"}
	ErrPreparingOracleData        = AppError{1014, "failed to prepare oracle data"}
	ErrMessageHashing             = AppError{1015, "failed to prepare hash message"}
	ErrPreparingProofData         = AppError{1016, "failed to prepare proof data"}
	ErrFormattingProofData        = AppError{1017, "failed to format proof data"}
	ErrGeneratingAttestationHash  = AppError{1018, "failed to generate to attestation hash"}
	ErrPreparingEncodedProof      = AppError{1019, "failed to prepare encoded request proof"}
	ErrFormattingEncodedProofData = AppError{1020, "failed to format encoded proof data"}
	ErrUserDataTooShort           = AppError{1021, "userData too short for expected zeroing"}
	ErrCreatingRequestHash        = AppError{1022, "failed to create request  hash"}
	ErrCreatingTimestampedHash    = AppError{1023, "failed to create timestamped  hash"}
	ErrFormattingQuote            = AppError{1024, "failed to format quote"}
	ErrReportHashing              = AppError{1025, "failed to hash the oracle report"}
	ErrGeneratingSignature        = AppError{1026, "failed to generate signature"}
	ErrFetchingData               = AppError{1027, "failed to fetch the data from the provided endpoint"}
	ErrReadingHTMLContent         = AppError{1028, "failed to read HTML content"}
	ErrParsingHTMLContent         = AppError{1029, "failed to parse HTML content"}
	ErrSelectorNotFound           = AppError{1030, "selector not found"}
	ErrInvalidMap                 = AppError{1031, "expected map but got something else"}
	ErrKeyNotFound                = AppError{1032, "selector not found"}
	ErrJSONDecoding               = AppError{1033, "failed to decode the JSON response"}
	ErrJSONEncoding               = AppError{1034, "failed to encode data to JSON"}
	ErrReadingTargetInfo          = AppError{1035, "failed to read the target info"}
	ErrWrittingReportData         = AppError{1036, "failed to write the report data"}
	ErrGeneratingQuote            = AppError{1037, "failed to generate the quote"}
	ErrReadingQuote               = AppError{1038, "failed to read the quote"}
	ErrWrittingTargetInfo         = AppError{1039, "failed to write the target info"}
	ErrWrappingQuote              = AppError{1040, "failed to wrap quote in openenclave format"}
	ErrEncodingAttestationData    = AppError{1041, "failed to encode attestation data"}
	ErrEncodingResponseFormat     = AppError{1042, "failed to encode response format"}
	ErrEncodingEncodingOptions    = AppError{1043, "failed to encode encoding options"}
	ErrEncodingHeaders            = AppError{1044, "failed to encode headers"}
	ErrEncodingOptionalFields     = AppError{1045, "failed to encode optional fields"}
	ErrPreparationCriticalError   = AppError{1046, "critical error while preparing data"}
	ErrWrittingAttestationData    = AppError{1047, "failed to write attestation data to buffer"}
	ErrWrittingTimestamp          = AppError{1048, "failed to write timestamp to buffer"}
	ErrWrittingStatusCode         = AppError{1049, "failed to write status code to buffer"}
	ErrWrittingUrl                = AppError{1050, "failed to write url to buffer"}
	ErrWrittingSelector           = AppError{1051, "failed to write selector to buffer"}
	ErrWrittingResponseFormat     = AppError{1052, "failed to write response format to buffer"}
	ErrWrittingRequestMethod      = AppError{1053, "failed to write request method to buffer"}
	ErrWrittingEncodingOptions    = AppError{1054, "failed to write encoding options to buffer"}
	ErrWrittingRequestHeaders     = AppError{1055, "failed to write request headers to buffer"}
	ErrWrittingOptionalFields     = AppError{1056, "failed to write optinal headers to buffer"}
	ErrInvalidSelectorPart        = AppError{1057, "invalid selector part"}
	ErrExpectedArray              = AppError{1058, "expected array at key"}
	ErrIndexOutOfBound            = AppError{1059, "index out of bounds"}
	ErrMissingSymbol              = AppError{1060, "symbol parameter is required"}
	ErrInvalidSymbol              = AppError{1061, "invalid symbol. Supported symbols: BTC, ETH, ALEO"}
	ErrPriceFeedFailed            = AppError{1062, "failed to get price feed data"}
	ErrReadingReport              = AppError{1063, "failed to read the report"}
	ErrUnsupportedPriceFeedURL    = AppError{1064, "unsupported price feed URL"}
	// Exchange configuration errors
	ErrExchangeNotConfigured      = AppError{1065, "exchange not configured"}
	ErrSymbolNotSupportedByExchange = AppError{1066, "symbol not supported by exchange"}
	ErrUnsupportedExchange        = AppError{1067, "unsupported exchange"}
	// Exchange API errors
	ErrExchangeFetchFailed        = AppError{1068, "failed to fetch from exchange"}
	ErrExchangeInvalidStatusCode  = AppError{1069, "exchange returned invalid status code"}
	ErrExchangeResponseDecodeFailed = AppError{1070, "failed to decode exchange response"}
	ErrExchangeResponseParseFailed = AppError{1071, "failed to parse exchange response"}
	// Data format errors
	ErrInvalidPriceFormat         = AppError{1072, "invalid price format"}
	ErrInvalidVolumeFormat        = AppError{1073, "invalid volume format"}
	ErrInvalidExchangeResponseFormat = AppError{1074, "invalid exchange response format"}
	ErrInvalidDataFormat          = AppError{1075, "invalid data format"}
	ErrInvalidItemFormat          = AppError{1076, "invalid item format"}
	ErrNoDataInResponse           = AppError{1077, "no data in response"}
	// Data parsing errors
	ErrPriceParseFailed           = AppError{1078, "failed to parse price"}
	ErrVolumeParseFailed          = AppError{1079, "failed to parse volume"}
	// Price feed validation errors
	ErrInsufficientExchangeData   = AppError{1080, "insufficient data from exchanges"}
	// Encoding validation errors
	ErrInvalidEncodingPrecision   = AppError{1081, "invalid encoding precision"}
)
