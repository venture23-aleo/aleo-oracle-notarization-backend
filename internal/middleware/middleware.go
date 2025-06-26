package middleware

import (
	"log"
	"net/http"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// RequireJSONContentType is a middleware that checks if the request has the correct Content-Type
func RequireJSONContentType(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Check if the content type is not JSON.
		if req.Header.Get("Content-Type") != "application/json" {
			log.Printf("Invalid Content-Type: %s for %s %s", req.Header.Get("Content-Type"), req.Method, req.URL.Path)
			
			// Create error response
			appErr := appErrors.AppError{
				Code:    1001, // Invalid request data error code
				Message: "Content-Type must be application/json",
			}
			
			utils.WriteJsonError(w, http.StatusBadRequest, appErr, "")
			return
		}
		
		// Call the next handler
		next(w, req)
	}
} 