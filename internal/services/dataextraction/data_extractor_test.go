package data_extraction

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
)

func TestMakeHTTPRequest_ValidRequests(t *testing.T) {

	elementType := "element"
	requestContentType := "application/json"
	requestBody := `{"name":"john","age":25	}`
	// valueType := "value"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html><head><title>Hello, World!</title></head><body><h1>Google</h1><span>10000.245</span><p>1000</p><h2></h2></body></html>`))
		case "/json":
			w.WriteHeader(http.StatusOK)
		case "/json_post":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"price": "10000.245", "volume": "1000.00"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	testCases := []struct {
		name               string
		attestationRequest attestation.AttestationRequest
	}{
		{
			name: "json request with valid response",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "symbol",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
		},
		{
			name: "html request with valid url",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "symbol",
				HTMLResultType: &elementType,
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
		},
		{
			name: "json request with valid url and POST method",
			attestationRequest: attestation.AttestationRequest{
				Url:                server.URL + "/json_post",
				RequestMethod:      "POST",
				RequestContentType: &requestContentType,
				ResponseFormat:     "json",
				Selector:           "name",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
				RequestBody: &requestBody,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resp, err := makeHTTPRequestToTarget(context.Background(), testCase.attestationRequest)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		})
	}
}

func TestMakeHTTPRequest_InvalidRequests(t *testing.T) {

	requestContentType := "application/json"
	requestBody := `{"name":"john","age":25	}`
	// valueType := "value"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html_404":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
		case "/json_404":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
		case "/json_500":
			w.WriteHeader(http.StatusInternalServerError)
		case "/html_500":
			w.WriteHeader(http.StatusInternalServerError)
		default:
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
			name: "json request with 500 error",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json_500",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "symbol",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: appErrors.ErrFetchingData,
		},
		{
			name: "html request with 500 error",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html_500",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "symbol",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: appErrors.ErrFetchingData,
		},
		{
			name: "json request with 404 error",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/json_404",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "symbol",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: appErrors.ErrInvalidStatusCode.WithResponseStatusCode(http.StatusNotFound),
		},
		{
			name: "html request with 404 error",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html_404",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "symbol",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError: appErrors.ErrInvalidStatusCode.WithResponseStatusCode(http.StatusNotFound),
		},
		{
			name: "json request with post method and invalid url",
			attestationRequest: attestation.AttestationRequest{
				Url:                server.URL + "/json_post_invalid",
				RequestMethod:      "POST",
				RequestContentType: &requestContentType,
				ResponseFormat:     "json",
				Selector:           "name",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
				RequestBody: &requestBody,
			},
			expectedError: appErrors.ErrInvalidStatusCode.WithResponseStatusCode(http.StatusNotFound),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := makeHTTPRequestToTarget(context.Background(), testCase.attestationRequest)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}
func TestExtractDataFromTargetURL_ValidRequest(t *testing.T) {

	htmlResponse := "<html><head><title>Hello, World!</title></head><body><h1>Google</h1><span>10000.245</span><p>1000</p><h2></h2></body></html>"
	jsonResponse := `{"price": "10000.245", "volume": "1000.00"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(htmlResponse))
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jsonResponse))
		default:
		}
	}))
	defer server.Close()

	elementType := "element"
	valueType := "value"

	testCases := []struct {
		name                    string
		request                 attestation.AttestationRequest
		expectedAttestationData ExtractDataResult
		checkAttestationData    bool
	}{
		{
			name: "valid html request with element type",
			request: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "html/head/title",
				HTMLResultType: &elementType,
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedAttestationData: ExtractDataResult{
				ResponseBody:    htmlResponse,
				AttestationData: "<title>Hello, World!</title>",
				StatusCode:      http.StatusOK,
			},
			checkAttestationData: true,
		},
		{
			name: "valid html request with value type",
			request: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "html/head/title",
				HTMLResultType: &valueType,
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedAttestationData: ExtractDataResult{
				ResponseBody:    htmlResponse,
				AttestationData: "Hello, World!",
				StatusCode:      http.StatusOK,
			},
			checkAttestationData: true,
		},
		{
			name: "valid json request",
			request: attestation.AttestationRequest{
				Url:            server.URL + "/json",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "price",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 10,
				},
			},
			expectedAttestationData: ExtractDataResult{
				ResponseBody:    jsonResponse,
				AttestationData: "10000.2450000000",
				StatusCode:      http.StatusOK,
			},
			checkAttestationData: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := ExtractDataFromTargetURL(context.Background(), testCase.request, time.Now().Unix())
			assert.Nil(t, err)
			if testCase.checkAttestationData {
				assert.Equal(t, testCase.expectedAttestationData, result)
			}
		})
	}
}

func TestExtractDataFromTargetURL_InvalidRequest(t *testing.T) {

	htmlResponse := "<html><head><title>Hello, World!</title></head><body><h1>Google</h1><span>10000.245</span><p>1000</p><h2></h2></body></html>"
	jsonResponse := `{"price": "10000.245", "volume": "1000.00"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(htmlResponse))
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jsonResponse))
		default:
		}
	}))
	defer server.Close()

	testCases := []struct {
		name                    string
		request                 attestation.AttestationRequest
		expectedError           *appErrors.AppError
		expectedAttestationData ExtractDataResult
		checkAttestationData    bool
	}{
		{
			name: "invalid response format",
			request: attestation.AttestationRequest{
				Url:            server.URL + "/text",
				ResponseFormat: "text",
				RequestMethod:  "GET",
				Selector:       "price",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedError:           appErrors.ErrInvalidResponseFormat,
			expectedAttestationData: ExtractDataResult{},
			checkAttestationData:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := ExtractDataFromTargetURL(context.Background(), testCase.request, time.Now().Unix())
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err)
			if testCase.checkAttestationData {
				assert.Equal(t, testCase.expectedAttestationData, result)
			}
		})
	}
}

func TestFormatAttestationData_ValidData(t *testing.T) {

	testCases := []struct {
		name            string
		attestationData ExtractDataResult
		encodingOptions string
		expectedError   *appErrors.AppError
	}{
		{
			name: "valid float data",
			attestationData: ExtractDataResult{
				AttestationData: "10000.245",
			},
			encodingOptions: "float",
			expectedError:   nil,
		},
		{
			name: "valid string data",
			attestationData: ExtractDataResult{
				AttestationData: "quick brown fox jumps over the lazy dog",
			},
			encodingOptions: "string",
			expectedError:   nil,
		},
		{
			name: "valid int data",
			attestationData: ExtractDataResult{
				AttestationData: "10000",
			},
			encodingOptions: "int",
			expectedError:   nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := formatAttestationData(context.Background(), testCase.attestationData.AttestationData, testCase.encodingOptions, 0)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestFormatAttestationData_InvalidData(t *testing.T) {

	testCases := []struct {
		name            string
		attestationData ExtractDataResult
		encodingOptions string
		expectedError   *appErrors.AppError
	}{
		{
			name: "invalid float data",
			attestationData: ExtractDataResult{
				AttestationData: "invalid",
			},
			encodingOptions: "float",
			expectedError:   appErrors.ErrParsingFloatValue,
		},
		{
			name: "invalid int data",
			attestationData: ExtractDataResult{
				AttestationData: "invalid",
			},
			encodingOptions: "int",
			expectedError:   appErrors.ErrParsingIntValue,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := formatAttestationData(context.Background(), testCase.attestationData.AttestationData, testCase.encodingOptions, 0)
			assert.NotNil(t, err)
		})
	}
}
