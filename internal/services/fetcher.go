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

func getNestedValueFlexible(m map[string]interface{}, path string) (interface{}, error) {
	re := regexp.MustCompile(`^(\w+)?(?:\[(\d+)\])?$`)
	current := interface{}(m)

	normalized := strings.ReplaceAll(path, ".[", "[")

	for _, part := range strings.Split(normalized, ".") {
		matches := re.FindStringSubmatch(part)
		if len(matches) < 2 {
			return nil, appErrors.ErrInvalidSelectorPart
		}

		key := matches[1]
		indexStr := matches[2]

		// Access the map
		asMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, appErrors.ErrInvalidMap
		}

		value, exists := asMap[key]
		if !exists {
			return nil, appErrors.ErrKeyNotFound
		}

		// If index is specified, access the array element
		if indexStr != "" {
			array, ok := value.([]interface{})
			if !ok {
				log.Printf("expected array at '%s'", key)

				return nil, appErrors.ErrExpectedArray
			}
			idx, _ := strconv.Atoi(indexStr)
			if idx < 0 || idx >= len(array) {
				log.Printf("index %d out of bounds for '%s'", idx, key)
				return nil, appErrors.ErrIndexOutOfBound
			}
			current = array[idx]
		} else {
			current = value
		}
	}

	return current, nil
}

func FetchDataFromAPIEndpoint(attestationRequest AttestationRequest) (string, string, int, error) {

	var bodyReader io.Reader

	if attestationRequest.RequestBody != nil {
		bodyReader = strings.NewReader(*attestationRequest.RequestBody)
	}

	url := "https://" + attestationRequest.Url

	req, err := http.NewRequest(attestationRequest.RequestMethod, url, bodyReader)

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrInvalidHTTPRequest
	}

	for key, value := range attestationRequest.RequestHeaders {
		req.Header.Set(key, value)
	}

	if attestationRequest.RequestContentType != nil {
		req.Header.Set("Content-Type", *attestationRequest.RequestContentType)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return "", "", resp.StatusCode, appErrors.ErrFetchingData
	}

	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return "", "", resp.StatusCode, appErrors.ErrFetchingData
	}

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrFetchingData
	}

	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrJSONDecoding
	}

	value, err := getNestedValueFlexible(response, attestationRequest.Selector)

	if err != nil {
		return "", "", http.StatusInternalServerError, err
	}

	valueStr := fmt.Sprintf("%v", value)

	if attestationRequest.EncodingOptions.Value == "float" {
		splitted := strings.SplitN(valueStr, ".", 2) // Split at most once

		// If there's no decimal point, or it's already within limits
		if len(splitted) == 1 || len(splitted[1]) <= encoding.ENCODING_OPTION_FLOAT_MAX_PRECISION {
			// safe to use as-is
		} else {
			// Truncate the decimal part to max precision
			splitted[1] = splitted[1][:encoding.ENCODING_OPTION_FLOAT_MAX_PRECISION]
			valueStr = splitted[0] + "." + splitted[1]
		}
	}

	jsonBytes, err := json.Marshal(response)

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrJSONEncoding
	}

	// Convert to string
	jsonString := string(jsonBytes)

	return jsonString, valueStr, http.StatusOK, nil
}

func ScrapDataFromWebsite(attestationRequest AttestationRequest) (string, string, int, error) {
	var bodyReader io.Reader

	if attestationRequest.RequestBody != nil {
		bodyReader = strings.NewReader(*attestationRequest.RequestBody)
	}

	url := "https://" + attestationRequest.Url

	req, err := http.NewRequest(attestationRequest.RequestMethod, url, bodyReader)

	for key, value := range attestationRequest.RequestHeaders {
		req.Header.Set(key, value)
	}

	if attestationRequest.RequestContentType != nil {
		req.Header.Set("Content-Type", *attestationRequest.RequestContentType)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil || resp == nil {
		return "", "", http.StatusBadRequest, appErrors.ErrFetchingData
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return "", "", resp.StatusCode, appErrors.ErrReadingHTMLContent
	}

	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return "", "", resp.StatusCode, appErrors.ErrReadingHTMLContent
	}

	// Read the full HTML content
	htmlBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrReadingHTMLContent
	}

	htmlContent := string(htmlBytes)

	htmlDoc, err := htmlquery.Parse(strings.NewReader(htmlContent))

	if err != nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrParsingHTMLContent
	}

	result, err := htmlquery.Query(htmlDoc, attestationRequest.Selector)

	if err != nil || result == nil {
		return "", "", http.StatusInternalServerError, appErrors.ErrSelectorNotFound
	}

	log.Printf("Data: %v", result.FirstChild.Data)

	return htmlContent, result.FirstChild.Data, resp.StatusCode, nil
}

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

	if attestationRequest.ResponseFormat == "html" {
		responseBody, data, statusCode, err := ScrapDataFromWebsite(attestationRequest)

		if err != nil {
			return "", "", statusCode, err
		}

		return responseBody, data, statusCode, nil

	} else if attestationRequest.ResponseFormat == "json" {
		responseBody, data, statusCode, err := FetchDataFromAPIEndpoint(attestationRequest)

		if err != nil {
			return "", "", statusCode, err
		}
		return responseBody, data, statusCode, nil

	} else {
		return "", "", http.StatusNotFound, appErrors.ErrInvalidResponseFormat
	}
}
