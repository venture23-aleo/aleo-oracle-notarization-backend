package data_extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

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
	resp, err := makeHTTPRequest(ctx, attestationRequest)
	if err != nil {
		reqLogger.Error("Error making HTTP request", "error", err, "url", attestationRequest.Url)
		return ExtractDataResult{
			StatusCode: resp.StatusCode,
		}, err
	}
	defer resp.Body.Close()

	// Create the response.
	var response interface{}

	// Decode the response.
	if decodeErr := json.NewDecoder(resp.Body).Decode(&response); decodeErr != nil {
		reqLogger.Error("Error decoding JSON: ", "error", decodeErr)
		return ExtractDataResult{
			StatusCode: resp.StatusCode,
		}, appErrors.NewAppError(appErrors.ErrJSONDecoding)
	}


	// Marshal the response.
	jsonBytes, marshalErr := json.Marshal(response)

		// Convert the JSON bytes to a string.
	jsonString := string(jsonBytes)

	// Check if the error is not nil.
	if marshalErr != nil {
		reqLogger.Error("Error marshalling JSON: ", "error", marshalErr)
		return ExtractDataResult{
			StatusCode: http.StatusInternalServerError,
		}, appErrors.NewAppError(appErrors.ErrJSONEncoding)
	}

	value := gjson.Get(jsonString, normalizeSelector(attestationRequest.Selector))
	if !value.Exists() {
		reqLogger.Error("Error getting value from JSON: ", "error", attestationRequest.Selector)
		return ExtractDataResult{
			StatusCode: resp.StatusCode,
		}, appErrors.NewAppError(appErrors.ErrKeyNotFound)
	}

	valueStr := fmt.Sprintf("%v", value)

	// Apply float precision if needed
	if attestationRequest.EncodingOptions.Value == "float" {
		_, floatErr := strconv.ParseFloat(valueStr, 64)
		if floatErr != nil {
			reqLogger.Error("Error parsing float value: ", "error", floatErr)
			return ExtractDataResult{
				StatusCode: resp.StatusCode,
			}, appErrors.NewAppError(appErrors.ErrInvalidEncodingOption)
		}
		valueStr = applyFloatPrecision(valueStr, attestationRequest.EncodingOptions.Precision)
	}


	



	// Return the JSON string, value, and status code.
	return ExtractDataResult{
		ResponseBody:    jsonString,
		AttestationData: valueStr,
		StatusCode:      http.StatusOK,
	}, nil
}








func normalizeSelector(selector string) string {
	// Replace [n] with .n where n is a number
	re := regexp.MustCompile(`\.?\[(\w+)\]`)
	normalized := re.ReplaceAllString(selector, `.$1`)
	return normalized
}