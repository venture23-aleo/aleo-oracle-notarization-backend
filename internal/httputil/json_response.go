package http

import (
	"encoding/json"
	"net/http"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

// WriteJsonSuccess writes a JSON success response with optional message and data
func WriteJsonSuccess(w http.ResponseWriter, statusCode int, data interface{}) {

	// Set the content type.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Encode the data.
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", "error", err)
		WriteJsonError(w, http.StatusInternalServerError, appErrors.ErrJSONEncoding)
		return
	}
}

// WriteJsonError writes a JSON error response with message and error code
func WriteJsonError(w http.ResponseWriter, statusCode int, appError *appErrors.AppError) {

	if appError == nil {
		appError = appErrors.ErrInternal
	}

	// Set the content type.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	requestID := w.Header().Get("X-Request-ID")

	// Set the request ID.
	appErrorWithRequestID := appErrors.AppErrorWithRequestID{
		AppError:  *appError,
		RequestID: requestID,
	}

	// Encode the error response.
	json.NewEncoder(w).Encode(appErrorWithRequestID)
}
