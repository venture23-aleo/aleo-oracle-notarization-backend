package data_extraction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.InitLogger("DEBUG")

	// Run the tests
	m.Run()
}

func TestExtractDataFromHTML_WithValidRequest(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch {
		case strings.Contains(r.URL.Path, "html"):
			w.Header().Set("Content-Type", "text/html")
			response := "<html><head><title>Hello, World!</title></head><body><h1>Google</h1></body></html>"
			w.Write([]byte(response))
		default:
			// Return 404 for unknown endpoints
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	valueType := "value"
	elementType := "element"

	testCases := []struct {
		name               string
		attestationRequest attestation.AttestationRequest
		expectedPayload    ExtractDataResult
	}{
		{
			name: "valid html request with value type",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "/html/head/title",
				HTMLResultType: &valueType,
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    "<html><head><title>Hello, World!</title></head><body><h1>Google</h1></body></html>",
				AttestationData: "Hello, World!",
				StatusCode:      200,
			},
		},
		{
			name: "valid html request with element type",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "/html/head/title",
				HTMLResultType: &elementType,
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    "<html><head><title>Hello, World!</title></head><body><h1>Google</h1></body></html>",
				AttestationData: "<title>Hello, World!</title>",
				StatusCode:      200,
			},
		},
		{
			name: "valid html request with element type for body content",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "/html/body/h1",
				HTMLResultType: &elementType,
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    "<html><head><title>Hello, World!</title></head><body><h1>Google</h1></body></html>",
				AttestationData: "<h1>Google</h1>",
				StatusCode:      200,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := ExtractDataFromTargetURL(context.Background(), testCase.attestationRequest)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			t.Logf("Result: %v", result)

			// For JSON responses, compare the parsed structures
			if testCase.attestationRequest.ResponseFormat == "json" {
				assert.JSONEq(t, testCase.expectedPayload.ResponseBody, result.ResponseBody)
			} else {
				assert.Equal(t, testCase.expectedPayload.ResponseBody, result.ResponseBody)
			}

			assert.Equal(t, testCase.expectedPayload.AttestationData, result.AttestationData)
			assert.Equal(t, testCase.expectedPayload.StatusCode, result.StatusCode)
		})
	}
}

func TestExtractDataFromHTML_WithInvalidRequest(t *testing.T) {

	valueType := "value"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "html"):
			w.Header().Set("Content-Type", "text/html")
			response := "<html><head><title>Hello, World!</title></head><body><h1>Google</h1></body></html>"
			w.Write([]byte(response))
		default:
			// Return 404 for unknown endpoints
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	testCases := []struct {
		name               string
		attestationRequest attestation.AttestationRequest
		expectedPayload    *appErrors.AppError
	}{
		{
			name: "invalid url",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "/html/head/title",
				HTMLResultType: &valueType,
			},
			expectedPayload: appErrors.NewAppErrorWithResponseStatus(appErrors.ErrFetchingData, http.StatusNotFound),
		},
		{
			name: "invalid selector",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "html/head/titles",
				HTMLResultType: &valueType,
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrSelectorNotFound),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := ExtractDataFromTargetURL(context.Background(), testCase.attestationRequest)
			t.Logf("Error: %v", err)
			assert.Equal(t, testCase.expectedPayload, err)
		})
	}
}

func TestExtractDataFromJSON_WithValidRequest(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch {
		case strings.Contains(r.URL.Path, "json"):
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"symbol":      "BTCUSDT",
				"lastPrice":   "50000.00",
				"volume":      "1000.50",
				"priceChange": "100.00",
				"tokens":      []string{"BTC", "ETH", "ALEO"},
				"singleChars": []string{"A", "B", "C"},
			}
			json.NewEncoder(w).Encode(response)
		default:
			// Return 404 for unknown endpoints
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	testCases := []struct {
		name               string
		attestationRequest attestation.AttestationRequest
		expectedPayload    ExtractDataResult
	}{
		{
			name: "valid json request",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "lastPrice",
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    "{\"symbol\":\"BTCUSDT\",\"lastPrice\":\"50000.00\",\"volume\":\"1000.50\",\"priceChange\":\"100.00\",\"tokens\":[\"BTC\",\"ETH\",\"ALEO\"],\"singleChars\":[\"A\",\"B\",\"C\"]}",
				AttestationData: "50000.00",
				StatusCode:      200,
			},
		},
		{
			name: "valid json request with nested selector",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "tokens[0]",
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    "{\"symbol\":\"BTCUSDT\",\"lastPrice\":\"50000.00\",\"volume\":\"1000.50\",\"priceChange\":\"100.00\",\"tokens\":[\"BTC\",\"ETH\",\"ALEO\"],\"singleChars\":[\"A\",\"B\",\"C\"]}",
				AttestationData: "BTC",
				StatusCode:      200,
			},
		},
		{
			name: "valid json request with nested selector for single chars",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "singleChars[1]",
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    "{\"symbol\":\"BTCUSDT\",\"lastPrice\":\"50000.00\",\"volume\":\"1000.50\",\"priceChange\":\"100.00\",\"tokens\":[\"BTC\",\"ETH\",\"ALEO\"],\"singleChars\":[\"A\",\"B\",\"C\"]}",
				AttestationData: "B",
				StatusCode:      200,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := ExtractDataFromTargetURL(context.Background(), testCase.attestationRequest)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			t.Logf("Result: %v", result)

			// For JSON responses, compare the parsed structures
			if testCase.attestationRequest.ResponseFormat == "json" {
				assert.JSONEq(t, testCase.expectedPayload.ResponseBody, result.ResponseBody)
			} else {
				assert.Equal(t, testCase.expectedPayload.ResponseBody, result.ResponseBody)
			}

			assert.Equal(t, testCase.expectedPayload.AttestationData, result.AttestationData)
			assert.Equal(t, testCase.expectedPayload.StatusCode, result.StatusCode)
		})
	}
}

func TestExtractDataFromJSON_WithInvalidRequest(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch {
		case strings.Contains(r.URL.Path, "json"):
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"symbol":      "BTCUSDT",
				"lastPrice":   "50000.00",
				"volume":      "1000.50",
				"priceChange": "100.00",
				"tokens":      []string{"BTC", "ETH", "ALEO"},
				"singleChars": []string{"A", "B", "C"},
			}
			json.NewEncoder(w).Encode(response)
		default:
			// Return 404 for unknown endpoints
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	testCases := []struct {
		name               string
		attestationRequest attestation.AttestationRequest
		expectedPayload    *appErrors.AppError
	}{
		{
			name: "invalid url",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/invalid",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "lastPrice",
			},
			expectedPayload: appErrors.NewAppErrorWithResponseStatus(appErrors.ErrFetchingData, http.StatusNotFound),
		},
		{
			name: "invalid selector",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "invalidKey",
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrKeyNotFound),
		},
		{
			name: "invalid array index",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "tokens[10]",
			},
			expectedPayload: appErrors.NewAppError(appErrors.ErrKeyNotFound),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := ExtractDataFromTargetURL(context.Background(), testCase.attestationRequest)
			t.Logf("Error: %v", err)
			assert.Equal(t, testCase.expectedPayload, err)
		})
	}
}
