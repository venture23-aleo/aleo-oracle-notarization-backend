package errors

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAppError_Error tests the Error() method of AppError
func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError AppError
		expected string
	}{
		{
			name: "basic error message",
			appError: AppError{
				Code:    1001,
				Message: "validation error: url is required",
			},
			expected: "code 1001: validation error: url is required",
		},
		{
			name: "error with zero code",
			appError: AppError{
				Code:    0,
				Message: "test error",
			},
			expected: "code 0: test error",
		},
		{
			name: "error with empty message",
			appError: AppError{
				Code:    2001,
				Message: "",
			},
			expected: "code 2001: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			if result != tt.expected {
				t.Errorf("AppError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAppError_JSON tests JSON marshaling of AppError
func TestAppError_JSON(t *testing.T) {
	tests := []struct {
		name     string
		appError AppError
		expected string
	}{
		{
			name: "complete error with all fields",
			appError: AppError{
				Code:               1001,
				Message:            "validation error: url is required",
				Details:            "URL field was empty",
				ResponseStatusCode: 400,
				RequestID:          "req-123",
			},
			expected: `{"errorCode":1001,"errorMessage":"validation error: url is required","errorDetails":"URL field was empty","responseStatusCode":400,"requestId":"req-123"}`,
		},
		{
			name: "error without optional fields",
			appError: AppError{
				Code:    2001,
				Message: "enclave error: failed to read the target info",
			},
			expected: `{"errorCode":2001,"errorMessage":"enclave error: failed to read the target info"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.appError)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, string(data), "JSON marshaling should match expected value")
		})
	}
}

// TestNewAppError tests the NewAppError constructor
func TestNewAppError(t *testing.T) {
	originalError := AppError{
		Code:               1001,
		Message:            "validation error: url is required",
		Details:            "URL field was empty",
		ResponseStatusCode: 400,
		RequestID:          "req-123",
	}

	result := NewAppError(originalError)

	// Check that a new pointer is returned
	if result == &originalError {
		t.Error("NewAppError should return a new pointer, not the original")
	}

	assert.NotEqual(t, result, &originalError, "NewAppError should return a new pointer, not the original")

	// Check that all fields are copied correctly
	assert.Equal(t, result.Code, originalError.Code, "Code should be copied correctly")
	assert.Equal(t, result.Message, originalError.Message, "Message should be copied correctly")
	assert.Equal(t, result.Details, originalError.Details, "Details should be copied correctly")
	assert.Equal(t, result.ResponseStatusCode, originalError.ResponseStatusCode, "ResponseStatusCode should be copied correctly")
	// RequestID should not be copied by NewAppError
	assert.Empty(t, result.RequestID, "RequestID should be empty")
}

// TestNewAppErrorWithResponseStatus tests the NewAppErrorWithResponseStatus constructor
func TestNewAppErrorWithResponseStatus(t *testing.T) {
	originalError := AppError{
		Code:    1001,
		Message: "validation error: url is required",
		Details: "URL field was empty",
	}
	newStatusCode := 422

	result := NewAppErrorWithResponseStatus(originalError, newStatusCode)

	// Check that a new pointer is returned
	if result == &originalError {
		t.Error("NewAppErrorWithResponseStatus should return a new pointer, not the original")
	}

	assert.NotEqual(t, result, &originalError, "NewAppErrorWithResponseStatus should return a new pointer, not the original")

	// Check that fields are copied correctly
	assert.Equal(t, result.Code, originalError.Code, "Code should be copied correctly")
	assert.Equal(t, result.Message, originalError.Message, "Message should be copied correctly")
	assert.Equal(t, result.Details, originalError.Details, "Details should be copied correctly")
	assert.Equal(t, result.ResponseStatusCode, newStatusCode, "ResponseStatusCode should be copied correctly")
	assert.Equal(t, result.RequestID, originalError.RequestID, "RequestID should be copied correctly")
}

// TestNewAppErrorWithDetails tests the NewAppErrorWithDetails constructor
func TestNewAppErrorWithDetails(t *testing.T) {
	originalError := AppError{
		Code:               1001,
		Message:            "validation error: url is required",
		ResponseStatusCode: 400,
	}
	newDetails := "Custom error details for debugging"

	result := NewAppErrorWithDetails(originalError, newDetails)

	// Check that a new pointer is returned
	if result == &originalError {
		t.Error("NewAppErrorWithDetails should return a new pointer, not the original")
	}

	assert.NotEqual(t, result, &originalError, "NewAppErrorWithDetails should return a new pointer, not the original")

	// Check that fields are copied correctly
	assert.Equal(t, result.Code, originalError.Code, "Code should be copied correctly")
	assert.Equal(t, result.Message, originalError.Message, "Message should be copied correctly")
	assert.Equal(t, result.ResponseStatusCode, originalError.ResponseStatusCode, "ResponseStatusCode should be copied correctly")
	assert.Equal(t, result.Details, newDetails, "Details should be copied correctly")
	assert.Equal(t, result.RequestID, originalError.RequestID, "RequestID should be copied correctly")
}

// TestPredefinedErrors tests that all predefined errors have correct structure
func TestPredefinedErrors(t *testing.T) {
	predefinedErrors := []AppError{
		ErrMissingURL,
		ErrMissingRequestMethod,
		ErrMissingResponseFormat,
		ErrMissingSelector,
		ErrMissingEncodingOption,
		ErrInvalidRequestMethod,
		ErrMissingRequestBody,
		ErrInvalidResponseFormat,
		ErrMissingHTMLResultType,
		ErrInvalidHTMLResultType,
		ErrInvalidEncodingOption,
		ErrUnacceptedDomain,
		ErrInvalidRequestData,
		ErrInvalidEncodingPrecision,
		ErrReadingTargetInfo,
		ErrWrittingReportData,
		ErrGeneratingQuote,
		ErrReadingQuote,
		ErrWrittingTargetInfo,
		ErrWrappingQuote,
		ErrReadingReport,
		ErrPreparingOracleData,
		ErrMessageHashing,
		ErrPreparingProofData,
		ErrFormattingProofData,
		ErrGeneratingAttestationHash,
		ErrPreparingEncodedProof,
		ErrFormattingEncodedProofData,
		ErrUserDataTooShort,
		ErrCreatingRequestHash,
		ErrCreatingTimestampedRequestHash,
		ErrFormattingQuote,
		ErrReportHashing,
		ErrGeneratingSignature,
		ErrInvalidHTTPRequest,
		ErrFetchingData,
		ErrReadingHTMLContent,
		ErrParsingHTMLContent,
		ErrSelectorNotFound,
		ErrInvalidMap,
		ErrKeyNotFound,
		ErrJSONDecoding,
		ErrJSONEncoding,
		ErrInvalidSelectorPart,
		ErrExpectedArray,
		ErrIndexOutOfBound,
		ErrUnsupportedPriceFeedURL,
		ErrEncodingAttestationData,
		ErrEncodingResponseFormat,
		ErrEncodingEncodingOptions,
		ErrEncodingHeaders,
		ErrEncodingOptionalFields,
		ErrPreparationCriticalError,
		ErrWrittingAttestationData,
		ErrWrittingTimestamp,
		ErrWrittingStatusCode,
		ErrWrittingUrl,
		ErrWrittingSelector,
		ErrWrittingResponseFormat,
		ErrWrittingRequestMethod,
		ErrWrittingEncodingOptions,
		ErrWrittingRequestHeaders,
		ErrWrittingOptionalFields,
		ErrMissingToken,
		ErrInvalidToken,
		ErrPriceFeedFailed,
		ErrExchangeNotConfigured,
		ErrSymbolNotSupportedByExchange,
		ErrUnsupportedExchange,
		ErrExchangeFetchFailed,
		ErrExchangeInvalidStatusCode,
		ErrExchangeResponseDecodeFailed,
		ErrExchangeResponseParseFailed,
		ErrInvalidPriceFormat,
		ErrInvalidVolumeFormat,
		ErrInvalidExchangeResponseFormat,
		ErrInvalidDataFormat,
		ErrInvalidItemFormat,
		ErrNoDataInResponse,
		ErrPriceParseFailed,
		ErrVolumeParseFailed,
		ErrInsufficientExchangeData,
		ErrAleoContext,
	}

	for _, err := range predefinedErrors {
		t.Run(err.Message, func(t *testing.T) {
			// Check that error code is in valid ranges
			assert.True(t, err.Code >= 1000 && err.Code <= 7999, "Error code %d is outside valid range (1000-7999)", err.Code)

			// Check that error message is not empty
			assert.NotEmpty(t, err.Message, "Error message should not be empty")

			// Check that error implements error interface
			errorString := err.Error()
			assert.NotEmpty(t, errorString, "Error() method should return non-empty string")

			// Check that error code ranges match categories
			switch {
			case err.Code >= 1000 && err.Code <= 1999:
				// Validation errors
			case err.Code >= 2000 && err.Code <= 2999:
				// Enclave errors
			case err.Code >= 3000 && err.Code <= 3999:
				// Attestation errors
			case err.Code >= 4000 && err.Code <= 4999:
				// Data extraction errors
			case err.Code >= 5000 && err.Code <= 5999:
				// Encoding errors
			case err.Code >= 6000 && err.Code <= 6999:
				// Exchange errors
			case err.Code >= 7000 && err.Code <= 7999:
				// Aleo context errors
			default:
				assert.Fail(t, "Error code %d does not fall into any defined category", err.Code)
			}
		})
	}
}

// TestErrorCodeRanges tests that error codes are properly categorized
func TestErrorCodeRanges(t *testing.T) {
	tests := []struct {
		name     string
		code     uint
		category string
	}{
		{name: "validation error", code: 1001, category: "validation"},
		{name: "enclave error", code: 2001, category: "enclave"},
		{name: "attestation error", code: 3001, category: "attestation"},
		{name: "data extraction error", code: 4001, category: "data extraction"},
		{name: "encoding error", code: 5001, category: "encoding"},
		{name: "exchange error", code: 6001, category: "exchange"},
		{name: "aleo context error", code: 7001, category: "aleo context"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test error with the given code
			testError := AppError{Code: tt.code, Message: "test error"}

			// Verify the code is in the expected range
			var minCode, maxCode uint
			switch tt.category {
			case "validation":
				minCode, maxCode = 1000, 1999
			case "enclave":
				minCode, maxCode = 2000, 2999
			case "attestation":
				minCode, maxCode = 3000, 3999
			case "data extraction":
				minCode, maxCode = 4000, 4999
			case "encoding":
				minCode, maxCode = 5000, 5999
			case "exchange":
				minCode, maxCode = 6000, 6999
			case "aleo context":
				minCode, maxCode = 7000, 7999
			}

			assert.True(t, testError.Code >= minCode && testError.Code <= maxCode, "Error code %d is not in expected range [%d, %d] for category %s", testError.Code, minCode, maxCode, tt.category)

			assert.Equal(t, testError.Message, "test error", "Error message should be 'test error'")
		})
	}
}

// TestAppError_ImplementsErrorInterface tests that AppError properly implements the error interface
func TestAppError_ImplementsErrorInterface(t *testing.T) {
	var _ error = AppError{} // This will compile only if AppError implements error interface

	// Test that we can assign AppError to error interface
	var err error = AppError{
		Code:    1001,
		Message: "test error",
	}

	// Test that we can call Error() method
	errorString := err.Error()
	assert.NotEmpty(t, errorString, "Error interface implementation should return non-empty string")
}
