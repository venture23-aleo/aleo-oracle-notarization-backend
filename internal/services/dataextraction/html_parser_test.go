package data_extraction

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
)

func TestExtractDataFromHTML_WithValidRequest(t *testing.T) {

	response := "<html><head><title>Hello, World!</title></head><body><h1>Google</h1><span>10000.245</span><p>1000</p><h2></h2></body></html>"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch {
		case strings.Contains(r.URL.Path, "html"):
			w.Header().Set("Content-Type", "text/html")
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
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    response,
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
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    response,
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
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    response,
				AttestationData: "<h1>Google</h1>",
				StatusCode:      200,
			},
		},
		{
			name: "valid html request with float encoding option",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "/html/body/span",
				HTMLResultType: &valueType,
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 5,
				},
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    response,
				AttestationData: "10000.24500",
				StatusCode:      200,
			},
		},
		{
			name: "valid html request with float encoding option",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "/html/body/p",
				HTMLResultType: &valueType,
				EncodingOptions: encoding.EncodingOptions{
					Value: "int",
				},
			},
			expectedPayload: ExtractDataResult{
				ResponseBody:    response,
				AttestationData: "1000",
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

func TestExtractDataFromHTML_WithInvalidRequest(t *testing.T) {

	valueType := "value"
	// elementType := "element"

	response := "<html><head><title>Hello, World!</title></head><body><h1>Google</h1><span>10000.245</span><p>1000</p><h2></h2></body></html>"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "html"):
			w.Header().Set("Content-Type", "text/html")
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
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedPayload: appErrors.ErrInvalidStatusCode.WithResponseStatusCode(http.StatusNotFound),
		},
		{
			name: "invalid domain",
			attestationRequest: attestation.AttestationRequest{
				Url:            "test",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "html/head/titles",
				HTMLResultType: &valueType,
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedPayload: appErrors.ErrInvalidURL,
		},
		{
			name: "invalid selector",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "html/head/titles",
				HTMLResultType: &valueType,
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedPayload: appErrors.ErrSelectorNotFound,
		},
		{
			name: "empty value",
			attestationRequest: attestation.AttestationRequest{
				Url:            server.URL + "/html",
				RequestMethod:  "GET",
				ResponseFormat: "html",
				Selector:       "/html/body/h2",
				HTMLResultType: &valueType,
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			expectedPayload: appErrors.ErrEmptyAttestationData,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := ExtractDataFromTargetURL(context.Background(), testCase.attestationRequest, 0)
			assert.Equal(t, testCase.expectedPayload, err)
		})
	}
}
