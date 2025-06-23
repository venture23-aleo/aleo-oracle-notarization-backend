package utils

import (
	"encoding/json"
	"net/http"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

// Error response structure
type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    int    `json:"errorCode"`
	RequestID    string `json:"requestId,omitempty"`
}

// WriteJsonSuccess writes a JSON success response with optional message and data
func WriteJsonSuccess(w http.ResponseWriter, statusCode int, data interface{}) {

	// Set the content type.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Encode the data.
	json.NewEncoder(w).Encode(data)
}

// WriteJsonError writes a JSON error response with message and error code
func WriteJsonError(w http.ResponseWriter, statusCode int, appError appErrors.AppError, requestID string) {

	// Set the content type.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Create the error response.
	resp := ErrorResponse{
		ErrorMessage: appError.Message,
		ErrorCode:    appError.Code,
		RequestID:    requestID,
	}

	// Encode the error response.
	json.NewEncoder(w).Encode(resp)
}
