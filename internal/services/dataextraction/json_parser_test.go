package data_extraction

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
)

func TestExtractDataFromJSON_WithValidRequest(t *testing.T) {

	requestBody := `{"name":"john","age":25}`
	requestContentType := "application/json"
	serverResponse := `{"symbol":"BTCUSDT","lastPrice":"50000.00","volume":"1000.50","priceChange":"100.00","tokenId":100,"tokens":["BTC","ETH","ALEO"],"singleChars":["A","B","C"],"emptyValue":""}`
	serverResponsePost := `{"name":"john","age":25}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch {
		case strings.Contains(r.URL.Path, "json") && r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(serverResponse))
		case strings.Contains(r.URL.Path, "json") && r.Method == "POST":
			w.Header().Set("Content-Type", "application/json")
			bodyBytes, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			assert.Equal(t, requestBody, string(bodyBytes))
			w.WriteHeader(http.StatusOK)
			w.Write(bodyBytes)
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
			name: "valid json request with float encoding option",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "lastPrice",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    serverResponse,
				AttestationData: "50000.000000",
				StatusCode:      200,
			},
		},
		{
			name: "valid json request with int encoding option",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "tokenId",
				EncodingOptions: encoding.EncodingOptions{
					Value: "int",
				},
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    serverResponse,
				AttestationData: "100",
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
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    serverResponse,
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
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    serverResponse,
				AttestationData: "B",
				StatusCode:      200,
			},
		},
		{
			name: "valid json request with post method",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "POST",
				ResponseFormat: "json",
				Selector:       "name",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
				RequestBody:        &requestBody,
				RequestContentType: &requestContentType,
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    serverResponsePost,
				AttestationData: "john",
				StatusCode:      200,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := ExtractDataFromTargetURL(context.Background(), testCase.attestationRequest, 0)
			assert.Nil(t, err)

			// For JSON responses, compare the parsed structures
			if testCase.attestationRequest.ResponseFormat == "json" {
				assert.JSONEq(t, testCase.expectedPayload.ResponseBody, result.ResponseBody)
			} else {
				assert.Equal(t, testCase.expectedPayload.ResponseBody, result.ResponseBody)
			}

			assert.Equal(t, testCase.expectedPayload.AttestationData, result.AttestationData)
			assert.Equal(t, testCase.expectedPayload.StatusCode, result.StatusCode)
			_, err = formatAttestationData(context.Background(), result.AttestationData, testCase.attestationRequest.EncodingOptions.Value, testCase.attestationRequest.EncodingOptions.Precision)
			assert.Nil(t, err)
		})
	}
}

func TestExtractDataFromJSON_WithInvalidRequest(t *testing.T) {

	serverResponse := `{"symbol":"BTCUSDT","lastPrice":"50000.00","volume":"1000.50","priceChange":"100.00","tokenId":100,"tokens":["BTC","ETH","ALEO"],"singleChars":["A","B","C"],"emptyValue":""}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/text":
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Text Response"))
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(serverResponse))
		default:
			// Return 404 for unknown endpoints
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	testCases := []struct {
		name               string
		attestationRequest attestation.AttestationRequest
		expectedError      *appErrors.AppError
	}{
		{
			name: "invalid url",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/invalid",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "lastPrice",
			},
			expectedError: appErrors.ErrInvalidStatusCode.WithResponseStatusCode(http.StatusNotFound),
		},
		{
			name: "invalid selector",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "invalidKey",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: appErrors.ErrSelectorNotFound,
		},
		{
			name: "invalid array index",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "tokens[10]",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: appErrors.ErrSelectorNotFound,
		},
		{
			name: "text response for json format",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/text",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "symbol",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: appErrors.ErrDecodingJSONResponse,
		},
		{
			name: "empty value",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "emptyValue",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: appErrors.ErrEmptyAttestationData,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := ExtractDataFromJSON(context.Background(), testCase.attestationRequest)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}
