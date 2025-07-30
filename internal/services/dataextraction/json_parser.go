package data_extraction

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/tidwall/gjson"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
)

func normalizeJSONSelector(selector string) string {
	// Replace [n] with .n where n is a number
	re := regexp.MustCompile(`\.?\[(\w+)\]`)
	normalized := re.ReplaceAllString(selector, `.$1`)
	return normalized
}

// Package data_extraction provides data extraction capabilities for the Aleo Oracle Notarization Backend.
// This file contains JSON-specific data extraction functionality.

// ExtractDataFromJSON fetches and extracts data from JSON responses for attestation purposes.
//
// This function:
// 1. Makes an HTTP request to the specified URL
// 2. Parses the JSON response
// 3. Extracts a specific value using the provided selector (supports nested paths and array indexing)
// 4. Applies encoding options (e.g., float precision)
// 5. Returns the full response body and extracted data
//
// The selector supports dot notation and array indexing:
// - "data.price" - extracts price from data object
// - "items[0].value" - extracts value from first item in items array
// - "results.data[2].metrics.total" - complex nested path with array access
//
// Example usage:
//
//	request := services.AttestationRequest{
//	    Url: "https://api.example.com/price",
//	    Selector: "data.price",
//	    ResponseFormat: "json",
//	    EncodingOptions: encoding.EncodingOptions{Value: "float", Precision: 6}
//	}
//	result, err := ExtractDataFromJSON(request)
func ExtractDataFromJSON(ctx context.Context, attestationRequest attestation.AttestationRequest) (ExtractDataResult, *appErrors.AppError) {
	// Make the HTTP request
	reqLogger := logger.FromContext(ctx)
	resp, err := makeHTTPRequestToTarget(ctx, attestationRequest)
	if err != nil {
		reqLogger.Error("Error making HTTP request", "error", err, "url", attestationRequest.Url)
		return ExtractDataResult{}, err
	}
	defer resp.Body.Close()

	limitReader := io.LimitReader(resp.Body, constants.MaxResponseBodySize)

	bodyBytes, readErr := io.ReadAll(limitReader)
	if readErr != nil {
		reqLogger.Error("Error reading response body", "error", readErr)
		return ExtractDataResult{StatusCode: resp.StatusCode}, appErrors.ErrReadingJSONResponse
	}

	if !json.Valid(bodyBytes) {
		reqLogger.Error("Invalid JSON response", "body", string(bodyBytes))
		return ExtractDataResult{StatusCode: resp.StatusCode}, appErrors.ErrDecodingJSONResponse
	}

	value := gjson.GetBytes(bodyBytes, normalizeJSONSelector(attestationRequest.Selector))
	if !value.Exists() {
		reqLogger.Error("Key not found in JSON response", "key", attestationRequest.Selector)
		return ExtractDataResult{StatusCode: resp.StatusCode}, appErrors.ErrSelectorNotFound
	}

	valueStr := value.String()

	if valueStr == "" {
		return ExtractDataResult{}, appErrors.ErrEmptyAttestationData
	}

	formattedAttestationData, err := formatAttestationData(ctx, valueStr, attestationRequest.EncodingOptions.Value, attestationRequest.EncodingOptions.Precision)
	if err != nil {
		return ExtractDataResult{}, err
	}

	// Return the JSON string, value, and status code.
	return ExtractDataResult{
		ResponseBody:    string(bodyBytes),
		AttestationData: formattedAttestationData,
		StatusCode:      http.StatusOK,
	}, nil
}
