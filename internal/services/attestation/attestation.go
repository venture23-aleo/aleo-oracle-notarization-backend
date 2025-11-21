package attestation

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

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

}

// AttestationResponse is the response body for the attestation service.
type AttestationResponse struct {
	ReportType string `json:"reportType"` // The report type.

	AttestationRequest AttestationRequest `json:"attestationRequest"` // The attestation request.

	AttestationReport string `json:"attestationReport"` // The attestation report.

	AttestationTimestamp int64 `json:"timestamp"` // The attestation timestamp.

	AleoBlockHeight int64 `json:"aleoBlockHeight"` // The Aleo block height.

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

func (ar *AttestationRequest) Normalize() AttestationRequest {

	clone := *ar;

	clone.Url = strings.ToLower(strings.TrimSpace(clone.Url))
	clone.RequestMethod = strings.ToUpper(strings.TrimSpace(clone.RequestMethod))
	clone.ResponseFormat = strings.ToLower(strings.TrimSpace(clone.ResponseFormat))
	clone.EncodingOptions.Value = strings.ToLower(strings.TrimSpace(clone.EncodingOptions.Value))

	clone.RequestHeaders = make(map[string]string)

	for headerName, headerValue := range ar.RequestHeaders {	
		trimmedHeaderName := strings.ToLower(strings.TrimSpace(headerName))
		trimmedHeaderValue := strings.TrimSpace(headerValue)
		clone.RequestHeaders[trimmedHeaderName] = trimmedHeaderValue
	}

	if clone.HTMLResultType != nil {
		htmlResultType := strings.ToLower(strings.TrimSpace(*ar.HTMLResultType))
		clone.HTMLResultType = &htmlResultType
	}

	return clone
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
	if strings.HasPrefix(strings.ToLower(ar.Url), "http://") || strings.HasPrefix(strings.ToLower(ar.Url), "https://") {
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

	for key,value := range ar.RequestHeaders {
		if !isValidHeaderKey(key) {
			return appErrors.ErrInvalidHeaderKey
		}
		if !isValidHeaderValue(value) {
			return appErrors.ErrInvalidHeaderValue
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

var (
	// exclude control characters except for HT(0x09) for header value
	reControl = regexp.MustCompile(`[\x00-\x08\x0a-\x1f\x7f]`)

	// percent-encoded CR/LF with repeated %25 prefixes:
	rePercentNested = regexp.MustCompile(`(?i)%(?:25)*0[da]`)

	// unicode-style escapes: %u000d, \u000d, optionally with fewer zeros
	reUnicodeEscape = regexp.MustCompile(`(?i)(?:\\u0*0*(?:0d|0a)|%(?:25)*u0*0*(?:0d|0a))`)

	// encoded slash rn: %5Cr %255Cr, %255Cn, %255C%255Cr, %255C%255Cr%255Cn, %255C%255Cr%255Cn%255C%255Cr, %255C%255Cr%255Cn%255C%255Cr%255C%255Cn, %255C%255Cr%255Cn%255C%255Cr%255C%255Cn%255C%255Cr, %255C%255Cr%255Cn%255C%255Cr%255C%255Cn%255C%255Cr%255C%255Cn
	reEncodedSlashRN = regexp.MustCompile(`(?i)%(?:25)*5C[rn]`)

	// Single regex to match HTML character references for carriage return (&#13;, &#x0d;, &#x0D;) and line feed (&#10;, &#x0a;, &#x0A;)
	reHTMLCharRefCRLF = regexp.MustCompile(`(?i)&#(?:0*13|x0*0d|0*10|x0*0a);`)

	// header-name token per RFC 7230 tchar: ALPHA / DIGIT / "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~"
	reHeaderKey = regexp.MustCompile(`^[!#$%&'*+\-.^_` + "`" + `|~0-9A-Za-z]+$`)
)


// isValidHeaderKey returns true if the key is safe to use as an HTTP header key.
func isValidHeaderKey(key string) (bool) {
	// check if the key is empty
	if key == "" {
		return false
	}

	// check if the key contains CR or LF
	if strings.ContainsAny(key, "\r\n") {
		return false
	}

	// check if the key is a valid token
	if !reHeaderKey.MatchString(key) {
		return false
	}

	return true
}

// isValidHeaderValue returns true if the value is safe to use as an HTTP header value.
func isValidHeaderValue(value string) (bool) {
	const maxUnescapeRounds = 3
	cur := value

	for i := 0; i < maxUnescapeRounds; i++ {

		if strings.ContainsAny(cur, "\r\n") {
				return false
		}

		// Reject control characters
		if reControl.MatchString(cur) {
			return false
		}

		// Reject obfuscation patterns (percent-nested, unicode escapes, encoded slash RN, HTML char refs)
		if rePercentNested.MatchString(cur) ||
			reUnicodeEscape.MatchString(cur) ||
			reEncodedSlashRN.MatchString(cur) ||
			reHTMLCharRefCRLF.MatchString(cur) {
			return false
		}

		if !utf8.ValidString(cur) {
			return false
		}

		// Attempt URL unescape for next iteration
		unescaped, err := url.PathUnescape(cur)
		if err != nil || unescaped == cur {
			break // stop if unescape fails or nothing changed
		}
		cur = unescaped
	}

	return true
}
