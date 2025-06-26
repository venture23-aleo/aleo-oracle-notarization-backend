package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

	"github.com/antchfx/htmlquery"
)

// getNestedValueFlexible gets the nested value from the map.
func getNestedValueFlexible(m map[string]interface{}, path string) (interface{}, *appErrors.AppError) {

	// Create the regular expression.
	re := regexp.MustCompile(`^(\w+)?(?:\[(\d+)\])?$`)

	// Create the current value.
	current := interface{}(m)

	// Normalize the path.
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
			return nil, appErrors.NewAppError(appErrors.ErrInvalidMap)
		}

		// Get the value.
		value, exists := asMap[key]

		// Check if the value exists.
		if !exists {
			return nil, appErrors.NewAppError(appErrors.ErrKeyNotFound)
		}

		// If index is specified, access the array element.
		if indexStr != "" {

			// Check if the value is an array.
			array, ok := value.([]interface{})

			// Check if the array is valid.
			if !ok {
				log.Printf("expected array at '%s'", key)
				return nil, appErrors.NewAppError(appErrors.ErrExpectedArray)
			}

			// Convert the index to an integer.
			idx, _ := strconv.Atoi(indexStr)

			// Check if the index is out of bounds.
			if idx < 0 || idx >= len(array) {
				log.Printf("index %d out of bounds for '%s'", idx, key)
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

// ExtractDataFromJSON fetches the data from the JSON response.
func ExtractDataFromJSON(attestationRequest AttestationRequest) (string, string, int, *appErrors.AppError) {

	// Create the body reader.
	var bodyReader io.Reader

	// Check if the request body is not nil.
	if attestationRequest.RequestBody != nil {
		bodyReader = strings.NewReader(*attestationRequest.RequestBody)
	}

	// Create the URL.
	url := "https://" + attestationRequest.Url

	// Create the request.
	req, err := http.NewRequest(attestationRequest.RequestMethod, url, bodyReader)

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.NewAppError(appErrors.ErrInvalidHTTPRequest)
	}

	// Set the request headers.
	for key, value := range attestationRequest.RequestHeaders {
		req.Header.Set(key, value)
	}

	// Set the request content type.
	if attestationRequest.RequestContentType != nil {
		req.Header.Set("Content-Type", *attestationRequest.RequestContentType)
	}

	// Create the client.
	client := &http.Client{Timeout: 10 * time.Second}

	// Do the request.
	resp, err := client.Do(req)

	// Check if the error is not nil.
	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.NewAppError(appErrors.ErrFetchingData)
	}

	// Close the response body.
	defer resp.Body.Close()

	// Check if the status code is greater than or equal to 400 and less than 600.
	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		return "", "", resp.StatusCode, appErrors.NewAppErrorWithResponseStatus(appErrors.ErrFetchingData, resp.StatusCode)
	}

	// Create the response.
	var response map[string]interface{}

	// Decode the response.
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", http.StatusInternalServerError, appErrors.NewAppError(appErrors.ErrJSONDecoding)
	}

	// Get the value.
	value, parseErr := getNestedValueFlexible(response, attestationRequest.Selector)

	// Check if the error is not nil.
	if parseErr != nil {
		return "", "", http.StatusInternalServerError, parseErr
	}

	// Convert the value to a string.
	valueStr := fmt.Sprintf("%v", value)

	// Check if the encoding option is float.
	if attestationRequest.EncodingOptions.Value == "float" {
		precision := attestationRequest.EncodingOptions.Precision
		splitted := strings.SplitN(valueStr, ".", 2) // Split at most once

		// If there's no decimal point, or it's already within limits.
		if len(splitted) == 1 || len(splitted[1]) <= int(precision) {
			// Safe to use as-is.
		} else {
			// Truncate the decimal part to max precision.
			splitted[1] = splitted[1][:precision]
			valueStr = splitted[0] + "." + splitted[1]
		}
	}

	// Marshal the response.
	jsonBytes, err := json.Marshal(response)

	// Check if the error is not nil.
	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.NewAppError(appErrors.ErrJSONEncoding)
	}

	// Convert the JSON bytes to a string.
	jsonString := string(jsonBytes)

	// Return the JSON string, value, and status code.
	return jsonString, valueStr, http.StatusOK, nil
}

// ExtractDataFromHTML scrapes the data from the HTML response.
func ExtractDataFromHTML(attestationRequest AttestationRequest) (string, string, int, *appErrors.AppError) {

	// Create the body reader.
	var bodyReader io.Reader

	// Check if the request body is not nil.
	if attestationRequest.RequestBody != nil {
		bodyReader = strings.NewReader(*attestationRequest.RequestBody)
	}

	// Create the URL.
	url := "https://" + attestationRequest.Url

	// Create the request.
	req, err := http.NewRequest(attestationRequest.RequestMethod, url, bodyReader)

	if err != nil {
		return "", "", http.StatusBadRequest, appErrors.NewAppError(appErrors.ErrInvalidHTTPRequest)
	}

	// Set the request headers.
	for key, value := range attestationRequest.RequestHeaders {
		req.Header.Set(key, value)
	}

	// Set the request content type.
	if attestationRequest.RequestContentType != nil {
		req.Header.Set("Content-Type", *attestationRequest.RequestContentType)
	}

	// Create the client.
	client := &http.Client{}

	// Do the request.
	resp, err := client.Do(req)

	// Check if the error is not nil or the response is nil.
	if err != nil || resp == nil {
		return "", "", http.StatusBadRequest, appErrors.NewAppError(appErrors.ErrFetchingData)
	}

	// Close the response body.
	defer resp.Body.Close()

	// Check if the status code is greater than or equal to 400 and less than 600.
	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		return "", "", resp.StatusCode, appErrors.NewAppErrorWithResponseStatus(appErrors.ErrReadingHTMLContent, resp.StatusCode)
	}

	// Read the full HTML content.
	htmlBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.NewAppError(appErrors.ErrReadingHTMLContent)
	}

	// Convert the HTML bytes to a string.
	htmlContent := string(htmlBytes)

	// Parse the HTML content.
	htmlDoc, err := htmlquery.Parse(strings.NewReader(htmlContent))

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.NewAppError(appErrors.ErrParsingHTMLContent)
	}

	// Query the HTML content.
	result, err := htmlquery.Query(htmlDoc, attestationRequest.Selector)

	// Check if the error is not nil or the result is nil.
	if err != nil || result == nil {
		return "", "", http.StatusInternalServerError, appErrors.NewAppError(appErrors.ErrSelectorNotFound)
	}

	valueStr := result.FirstChild.Data

	if attestationRequest.EncodingOptions.Value == "float" {	
		valueStr1, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return "", "", http.StatusInternalServerError, appErrors.NewAppError(appErrors.ErrParsingHTMLContent)
		}
		valueStr = fmt.Sprintf("%.*f", attestationRequest.EncodingOptions.Precision, valueStr1)
	}

	// Return the HTML content, data, status code, and error.
	return htmlContent, valueStr, resp.StatusCode, nil
}

// ExtractPriceFeedData handles price feed requests and always returns the volume-weighted average price (VWAP)
// This ensures consistent and reliable price data for oracle attestations
func ExtractPriceFeedData(attestationRequest AttestationRequest) (string, string, int, *appErrors.AppError) {

	if attestationRequest.EncodingOptions.Value != "float" {
		return "", "", http.StatusBadRequest, appErrors.NewAppError(appErrors.ErrInvalidEncodingOption)
	}

	// Extract symbol from the price feed URL
	var symbol string	
	switch attestationRequest.Url {
	case constants.PriceFeedBtcUrl:
		symbol = "BTC"
	case constants.PriceFeedEthUrl:
		symbol = "ETH"
	case constants.PriceFeedAleoUrl:
		symbol = "ALEO"
	default:
		return "", "", http.StatusBadRequest, appErrors.NewAppError(appErrors.ErrUnsupportedPriceFeedURL)
	}

	// Get the price feed data
	result, appErr := GetPriceFeed(symbol)
	if appErr.Code != 0 {
		log.Printf("Error getting price feed for %s: %v", symbol, appErr)
		return "", "", http.StatusInternalServerError, appErr
	}

	
	// Marshal the response to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.NewAppError(appErrors.ErrJSONEncoding)
	}

	// Extract the value based on the selector
	// For price feeds, always use weightedAvgPrice (volume-weighted average price)
	var valueStr string = result.VolumeWeightedAvg
	
	// Handle precision based on encoding options

	precision := attestationRequest.EncodingOptions.Precision
	splitted := strings.SplitN(valueStr, ".", 2) // Split at most once

	// If there's no decimal point, or it's already within limits.
	if len(splitted) == 1 || len(splitted[1]) <= int(precision) {
		// Safe to use as-is.
	} else {
		// Truncate the decimal part to max precision.
		splitted[1] = splitted[1][:precision]
		valueStr = splitted[0] + "." + splitted[1]
	}

	return string(jsonBytes), valueStr, http.StatusOK, nil
}

// ExtractDataFromTargetURL fetches the data from the attestation request target URL.
func ExtractDataFromTargetURL(attestationRequest AttestationRequest) (string, string, int, *appErrors.AppError) {
	// Check if the URL is a price feed request
	if attestationRequest.Url == constants.PriceFeedBtcUrl || attestationRequest.Url == constants.PriceFeedEthUrl || attestationRequest.Url == constants.PriceFeedAleoUrl {
		return ExtractPriceFeedData(attestationRequest)
	} else if attestationRequest.ResponseFormat == "html" {
		return ExtractDataFromHTML(attestationRequest)
	} else if attestationRequest.ResponseFormat == "json" {
		return ExtractDataFromJSON(attestationRequest)
	} else {
		return "", "", http.StatusNotFound, appErrors.NewAppError(appErrors.ErrInvalidResponseFormat)
	}
}


