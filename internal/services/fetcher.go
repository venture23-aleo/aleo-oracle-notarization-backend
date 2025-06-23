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

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

	encoding "github.com/zkportal/aleo-oracle-encoding"

	"github.com/antchfx/htmlquery"
)

// getNestedValueFlexible gets the nested value from the map.
func getNestedValueFlexible(m map[string]interface{}, path string) (interface{}, error) {

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
			return nil, appErrors.ErrInvalidSelectorPart
		}

		// Get the key.
		key := matches[1]

		// Get the index.
		indexStr := matches[2]

		// Access the map.
		asMap, ok := current.(map[string]interface{})

		// Check if the map is valid.
		if !ok {
			return nil, appErrors.ErrInvalidMap
		}

		// Get the value.
		value, exists := asMap[key]

		// Check if the value exists.
		if !exists {
			return nil, appErrors.ErrKeyNotFound
		}

		// If index is specified, access the array element.
		if indexStr != "" {

			// Check if the value is an array.
			array, ok := value.([]interface{})

			// Check if the array is valid.
			if !ok {
				log.Printf("expected array at '%s'", key)
				return nil, appErrors.ErrExpectedArray
			}

			// Convert the index to an integer.
			idx, _ := strconv.Atoi(indexStr)

			// Check if the index is out of bounds.
			if idx < 0 || idx >= len(array) {
				log.Printf("index %d out of bounds for '%s'", idx, key)
				return nil, appErrors.ErrIndexOutOfBound
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

// FetchDataFromAPIEndpoint fetches the data from the API endpoint.
func FetchDataFromAPIEndpoint(attestationRequest AttestationRequest) (string, string, int, error) {

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
		return "", "", http.StatusInternalServerError, appErrors.ErrInvalidHTTPRequest
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

	// Check if the status code is greater than or equal to 400 and less than 500.
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return "", "", resp.StatusCode, appErrors.ErrFetchingData
	}

	// Check if the status code is greater than or equal to 500 and less than 600.
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return "", "", resp.StatusCode, appErrors.ErrFetchingData
	}

	// Check if the error is not nil.
	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrFetchingData
	}

	// Close the response body.
	defer resp.Body.Close()

	// Create the response.
	var response map[string]interface{}

	// Decode the response.
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrJSONDecoding
	}

	// Get the value.
	value, err := getNestedValueFlexible(response, attestationRequest.Selector)

	// Check if the error is not nil.
	if err != nil {
		return "", "", http.StatusInternalServerError, err
	}

	// Convert the value to a string.
	valueStr := fmt.Sprintf("%v", value)

	// Check if the encoding option is float.
	if attestationRequest.EncodingOptions.Value == "float" {
		splitted := strings.SplitN(valueStr, ".", 2) // Split at most once

		// If there's no decimal point, or it's already within limits.
		if len(splitted) == 1 || len(splitted[1]) <= encoding.ENCODING_OPTION_FLOAT_MAX_PRECISION {
			// Safe to use as-is.
		} else {
			// Truncate the decimal part to max precision.
			splitted[1] = splitted[1][:encoding.ENCODING_OPTION_FLOAT_MAX_PRECISION]
			valueStr = splitted[0] + "." + splitted[1]
		}
	}

	// Marshal the response.
	jsonBytes, err := json.Marshal(response)

	// Check if the error is not nil.
	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrJSONEncoding
	}

	// Convert the JSON bytes to a string.
	jsonString := string(jsonBytes)

	// Return the JSON string, value, and status code.
	return jsonString, valueStr, http.StatusOK, nil
}

// ScrapDataFromWebsite scrapes the data from the website.
func ScrapDataFromWebsite(attestationRequest AttestationRequest) (string, string, int, error) {

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
		return "", "", http.StatusBadRequest, appErrors.ErrFetchingData
	}

	// Close the response body.
	defer resp.Body.Close()

	// Check if the status code is greater than or equal to 400 and less than 500.
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return "", "", resp.StatusCode, appErrors.ErrReadingHTMLContent
	}

	// Check if the status code is greater than or equal to 500 and less than 600.
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return "", "", resp.StatusCode, appErrors.ErrReadingHTMLContent
	}

	// Read the full HTML content.
	htmlBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrReadingHTMLContent
	}

	// Convert the HTML bytes to a string.
	htmlContent := string(htmlBytes)

	// Parse the HTML content.
	htmlDoc, err := htmlquery.Parse(strings.NewReader(htmlContent))

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrParsingHTMLContent
	}

	// Query the HTML content.
	result, err := htmlquery.Query(htmlDoc, attestationRequest.Selector)

	// Check if the error is not nil or the result is nil.
	if err != nil || result == nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrSelectorNotFound
	}

	// Return the HTML content, data, status code, and error.
	return htmlContent, result.FirstChild.Data, resp.StatusCode, nil
}

// FetchDataFromAttestationRequest fetches the data from the attestation request.
func FetchDataFromAttestationRequest(attestationRequest AttestationRequest) (string, string, int, error) {
	// const attestationRequest = {
	//     url: 'google.com',
	//     requestMethod: 'GET',
	//     responseFormat: 'html',
	//     htmlResultType: 'value',
	//     selector: '/html/head/title',
	//     encodingOptions: {
	//         value: 'string',
	//     },
	// };

	// Check if the response format is HTML.
	if attestationRequest.ResponseFormat == "html" {
		responseBody, data, statusCode, err := ScrapDataFromWebsite(attestationRequest)

		// Check if the error is not nil.
		if err != nil {
			return "", "", statusCode, err
		}

		// Return the response body, data, status code, and error.
		return responseBody, data, statusCode, nil

		// Check if the response format is JSON.
	} else if attestationRequest.ResponseFormat == "json" {
		responseBody, data, statusCode, err := FetchDataFromAPIEndpoint(attestationRequest)

		// Check if the error is not nil.
		if err != nil {
			return "", "", statusCode, err
		}

		// Return the response body, data, status code, and error.
		return responseBody, data, statusCode, nil

		// Return the error.
	} else {
		return "", "", http.StatusNotFound, appErrors.ErrInvalidResponseFormat
	}
}
