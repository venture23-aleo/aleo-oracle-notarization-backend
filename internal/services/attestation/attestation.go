package attestation

import (
	"strings"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
)

// AttestationRequest is the request body for the attestation service.
type AttestationRequest struct {
	Url string `json:"url"` // The URL to fetch data from.

	RequestMethod  string  `json:"requestMethod"`            // The request method.
	Selector       string  `json:"selector,omitempty"`       // The selector.
	ResponseFormat string  `json:"responseFormat"`           // The response format.
	HTMLResultType *string `json:"htmlResultType,omitempty"` // The HTML result type.

	RequestBody        *string `json:"requestBody,omitempty"`        // The request body.
	RequestContentType *string `json:"requestContentType,omitempty"` // The request content type.

	RequestHeaders map[string]string `json:"requestHeaders,omitempty"` // The request headers.

	EncodingOptions encoding.EncodingOptions `json:"encodingOptions"` // The encoding options.

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

	ExtractedData string `json:"extractedData"` // The extracted data.
}

// Validate validates the attestation request and checks if target is whitelisted.
func (ar *AttestationRequest) Validate() *appErrors.AppError {

	// Check if the URL is empty.
	if ar.Url == "" {
		return appErrors.ErrMissingURL
	}

	// Check if the request method is empty.
	if ar.RequestMethod == "" {
		return appErrors.ErrMissingRequestMethod
	}

	// Check if the selector is empty.
	if ar.Selector == "" {
		return appErrors.ErrMissingSelector
	}

	if ar.ResponseFormat == "" {
		return appErrors.ErrMissingResponseFormat
	}

	// Check if the encoding option value is empty.
	if ar.EncodingOptions.Value == "" {
		return appErrors.ErrMissingEncodingOption
	}

	// Check if the request method is valid.
	if ar.RequestMethod != constants.RequestMethodGET && ar.RequestMethod != constants.RequestMethodPOST {
		return appErrors.ErrInvalidRequestMethod
	}

	// Check if the request body is required for POST requests.
	if ar.RequestMethod == constants.RequestMethodPOST && ar.RequestBody == nil {
		return appErrors.ErrMissingRequestBody
	}

	// Check if the request body is not allowed for GET requests.
	if ar.RequestMethod == constants.RequestMethodGET && ar.RequestBody != nil {
		return appErrors.ErrInvalidRequestBody
	}

	// Check if the request content type is not allowed for GET requests.
	if ar.RequestMethod == constants.RequestMethodGET && ar.RequestContentType != nil {
		return appErrors.ErrInvalidRequestContentType
	}

	// Check if the request content type is required for POST requests.
	if ar.RequestMethod == constants.RequestMethodPOST && ar.RequestContentType == nil {
		return appErrors.ErrMissingRequestContentType
	}

	// Check if the response format is valid.
	if ar.ResponseFormat != constants.ResponseFormatHTML && ar.ResponseFormat != constants.ResponseFormatJSON {
		return appErrors.ErrInvalidResponseFormat
	}

	// Check if the HTML result type is required for HTML response format.
	if ar.ResponseFormat == constants.ResponseFormatHTML && ar.HTMLResultType == nil {
		return appErrors.ErrMissingHTMLResultType
	}

	// Check if the HTML result type is valid.
	if ar.ResponseFormat == constants.ResponseFormatHTML && *ar.HTMLResultType != constants.HTMLResultTypeValue && *ar.HTMLResultType != constants.HTMLResultTypeElement {
		return appErrors.ErrInvalidHTMLResultType
	}

	// Check if the HTML result type is valid for the encoding option.
	if ar.ResponseFormat == constants.ResponseFormatHTML && *ar.HTMLResultType == constants.HTMLResultTypeElement && (ar.EncodingOptions.Value == constants.EncodingOptionInt || ar.EncodingOptions.Value == constants.EncodingOptionFloat) {
		return appErrors.ErrInvalidEncodingOptionForHTMLResultType
	}

	if ar.ResponseFormat == constants.ResponseFormatJSON && (ar.HTMLResultType != nil) {
		return appErrors.ErrInvalidHTMLResultTypeForJSONResponse
	}

	// Check if the encoding option is valid.
	if ar.EncodingOptions.Value != constants.EncodingOptionString && ar.EncodingOptions.Value != constants.EncodingOptionFloat && ar.EncodingOptions.Value != constants.EncodingOptionInt {
		return appErrors.ErrInvalidEncodingOption
	}

	// Check if the encoding option precision is not allowed for int or string encoding.
	if (ar.EncodingOptions.Value == constants.EncodingOptionInt || ar.EncodingOptions.Value == constants.EncodingOptionString) && (ar.EncodingOptions.Precision != 0) {
		return appErrors.ErrInvalidEncodingPrecision
	}

	// Check if the encoding option precision is valid (only for float encoding).
	if ar.EncodingOptions.Value == constants.EncodingOptionFloat && (ar.EncodingOptions.Precision == 0 || ar.EncodingOptions.Precision > encoding.ENCODING_OPTION_FLOAT_MAX_PRECISION) {
		return appErrors.ErrInvalidEncodingPrecision
	}

	// Check if the URL is invalid.
	if strings.HasPrefix(ar.Url, "http://") || strings.HasPrefix(ar.Url, "https://") {
		return appErrors.ErrInvalidTargetURL
	}

	// Check if the target is whitelisted.
	if !common.IsTargetWhitelisted(ar.Url) {
		return appErrors.ErrTargetNotWhitelisted
	}

	// Check if the request is a price feed request.
	if common.IsPriceFeedURL(ar.Url) {
		// Check if the encoding option is valid for price feed requests.
		if ar.EncodingOptions.Value != constants.EncodingOptionFloat {
			return appErrors.ErrInvalidEncodingOptionForPriceFeed
		}
		// Check if the response format is valid for price feed requests.
		if ar.ResponseFormat != constants.ResponseFormatJSON {
			return appErrors.ErrInvalidResponseFormatForPriceFeed
		}
		// Check if the request method is valid for price feed requests.
		if ar.RequestMethod != constants.RequestMethodGET {
			return appErrors.ErrInvalidRequestMethodForPriceFeed
		}
		// Check if the selector is valid for price feed requests.
		if ar.Selector != constants.PriceFeedSelector {
			return appErrors.ErrInvalidSelectorForPriceFeed
		}
	}

	return nil
}

// Masks unaccepted headers by replacing their values with "******"
func (ar *AttestationRequest) MaskUnacceptedHeaders() {
	finalHeaders := make(map[string]string)
	for headerName, headerValue := range ar.RequestHeaders {
		if !common.IsAcceptedHeader(headerName) {
			finalHeaders[headerName] = "******"
		} else {
			finalHeaders[headerName] = headerValue
		}
	}
	ar.RequestHeaders = finalHeaders
}
