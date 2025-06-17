package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
)

// Checks if a header name is in the list of allowed headers.
func isAcceptedHeader(header string) bool {
	for _, h := range constants.ALLOWED_HEADERS {
		if strings.EqualFold(h, header) {
			return true
		}
	}
	return false
}

// Masks unaccepted headers by replacing their values with "******"
func MaskUnacceptedHeaders(headers map[string]string) (map[string]string){
	finalHeaders := make(map[string]string)
	for headerName,headerValue := range headers {
		if !isAcceptedHeader(headerName) {
			finalHeaders[headerName] = "******"
		} else {
			finalHeaders[headerName] = headerValue
		}
	}
	return finalHeaders
}

func IsAcceptedDomain(endpoint string) (bool){
	parsedURL, err := url.Parse(fmt.Sprintf("https://%s",endpoint))
    if err != nil {
        fmt.Println("Error parsing URL:", err)
        return false
    }
	for _, domainName := range configs.WHITELISTED_DOMAINS {
		if domainName == parsedURL.Hostname() {
			return true
		}
	}
	return false
}

func ReverseBytes(b []byte) []byte {
	reversed := make([]byte, len(b))
	for i := range b {
		reversed[i] = b[len(b)-1-i]
	}
	return reversed
}

func GenerateShortRequestID() string {
    b := make([]byte, 16)
    _, err := rand.Read(b)
    if err != nil {
        return "unknown-request-id"
    }
    return hex.EncodeToString(b) // e.g., "f4e3d2a1b3c0d9e8"
}
