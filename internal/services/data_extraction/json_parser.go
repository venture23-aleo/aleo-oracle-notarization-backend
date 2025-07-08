package data_extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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
	var response map[string]interface{}

	// Decode the response.
	if decodeErr := json.NewDecoder(resp.Body).Decode(&response); decodeErr != nil {
		reqLogger.Error("Error decoding JSON: ", "error", decodeErr)
		return ExtractDataResult{
			StatusCode: resp.StatusCode,
		}, appErrors.NewAppError(appErrors.ErrJSONDecoding)
	}

	value, err := getNestedValue(response, attestationRequest.Selector)
	if err != nil {
		reqLogger.Error("Error getting nested value: ", "error", err)
		return ExtractDataResult{
			StatusCode: resp.StatusCode,
		}, err
	}

	valueStr := fmt.Sprintf("%v", value)

	// Apply float precision if needed
	if attestationRequest.EncodingOptions.Value == "float" {
		_, floatErr := strconv.ParseFloat(valueStr, 64)
		if floatErr != nil {
			reqLogger.Error("Error parsing float value: ", "error", floatErr)
			return ExtractDataResult{
				StatusCode: resp.StatusCode,
			}, appErrors.NewAppError(appErrors.ErrParsingHTMLContent)
		}
		valueStr = applyFloatPrecision(valueStr, attestationRequest.EncodingOptions.Precision)
	}

	// Marshal the response.
	jsonBytes, marshalErr := json.Marshal(response)

	// Check if the error is not nil.
	if marshalErr != nil {
		reqLogger.Error("Error marshalling JSON: ", "error", marshalErr)
		return ExtractDataResult{
			StatusCode: http.StatusInternalServerError,
		}, appErrors.NewAppError(appErrors.ErrJSONEncoding)
	}

	// Convert the JSON bytes to a string.
	jsonString := string(jsonBytes)

	// Return the JSON string, value, and status code.
	return ExtractDataResult{
		ResponseBody:    jsonString,
		AttestationData: valueStr,
		StatusCode:      http.StatusOK,
	}, nil
}

// getNestedValue extracts a value from a nested JSON structure using a flexible path selector.
//
// Supports:
// - Dot notation: "data.price"
// - Array indexing: "items[0].value"
// - Mixed paths: "results.data[2].metrics.total"
//
// The path is parsed using regex to handle both object keys and array indices.
// Returns the extracted value or an error if the path is invalid or the value doesn't exist.
//
// Examples:
//
//	getNestedValue(data, "price")           // data["price"]
//	getNestedValue(data, "items.[0].price") // data["items"][0]["price"]
//	getNestedValue(data, "items[0]")        // data["items"][0]
//	getNestedValue(data, "data.items[1].price") // data["data"]["items"][1]["price"]
func getNestedValue(m map[string]interface{}, path string) (interface{}, *appErrors.AppError) {
	// Create the regular expression for parsing path components
	// Matches: key[optional_index] where key is alphanumeric and index is numeric
	re := regexp.MustCompile(`^(\w+)?(?:\[(\d+)\])?$`)

	// Create the current value.
	current := interface{}(m)

	// Normalize the path by removing spaces around brackets
	normalized := strings.ReplaceAll(path, ".[", "[")

	// Split the path into parts.
	for _, part := range strings.Split(normalized, ".") {
		// Find the matches.
		matches := re.FindStringSubmatch(part)

		// Check if the matches are valid.
		if len(matches) < 2 {
			return nil, appErrors.NewAppError(appErrors.ErrInvalidSelectorPart)
		}

		// Get the key.
		key := matches[1]

		// Get the index.
		indexStr := matches[2]

		// Access the map.
		asMap, ok := current.(map[string]interface{})

		// Check if the map is valid.
		if !ok {
			logger.Error("Error getting nested value: Invalid map")
			return nil, appErrors.NewAppError(appErrors.ErrInvalidMap)
		}

		// Get the value.
		value, exists := asMap[key]

		// Check if the value exists.
		if !exists {
			logger.Error("Error getting nested value: Key not found")
			return nil, appErrors.NewAppError(appErrors.ErrKeyNotFound)
		}

		// If index is specified, access the array element.
		if indexStr != "" {
			// Check if the value is an array.
			array, ok := value.([]interface{})

			// Check if the array is valid.
			if !ok {
				logger.Error("Error getting nested value: Expected array")
				return nil, appErrors.NewAppError(appErrors.ErrExpectedArray)
			}

			// Convert the index to an integer.
			idx, _ := strconv.Atoi(indexStr)

			// Check if the index is out of bounds.
			if idx < 0 || idx >= len(array) {
				logger.Error("Error getting nested value: Index out of bounds")
				return nil, appErrors.NewAppError(appErrors.ErrIndexOutOfBound)
			}

			// Get the value.
			current = array[idx]
		} else {
			// Get the value.
			current = value
		}
	}

	// Return the value.
	return current, nil
}
