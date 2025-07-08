package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
)

// getPadding gets the padding for the array.
func GetPadding(arr []byte, alignment int) []byte {
	var paddingSize int
	overflow := len(arr) % alignment

	// Check if there is an overflow.
	if overflow != 0 {
		paddingSize = alignment - overflow // Calculate the padding size.
	} else {
		paddingSize = 0
	}

	// Create the padding.
	padding := make([]byte, paddingSize)

	// Return the padding.
	return padding
}

// padStringToLength pads the string to the target length.
func PadStringToLength(str string, paddingChar byte, targetLength int) string {

	// Pad the string to the target length.
	return str + strings.Repeat(string(paddingChar), targetLength-len(str))
}

// Checks if a header name is in the list of allowed headers.
func IsAcceptedHeader(header string) bool {
	for _, h := range constants.ALLOWED_HEADERS {
		if strings.EqualFold(h, header) {
			return true
		}
	}
	return false
}

// Masks unaccepted headers by replacing their values with "******"
func MaskUnacceptedHeaders(headers map[string]string) map[string]string {
	finalHeaders := make(map[string]string)
	for headerName, headerValue := range headers {
		if !IsAcceptedHeader(headerName) {
			finalHeaders[headerName] = "******"
		} else {
			finalHeaders[headerName] = headerValue
		}
	}
	return finalHeaders
}

// Checks if a domain is in the list of whitelisted domains.
func IsAcceptedDomain(endpoint string) bool {
	if endpoint == constants.PriceFeedBtcUrl || endpoint == constants.PriceFeedEthUrl || endpoint == constants.PriceFeedAleoUrl {
		return true
	}

	var urlToParse string
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		urlToParse = endpoint
	} else {
		urlToParse = fmt.Sprintf("https://%s", endpoint)
	}
	
	parsedURL, err := url.Parse(urlToParse)
	if err != nil {
		return false
	}
	for _, domainName := range configs.GetWhitelistedDomains() {
		if domainName == parsedURL.Hostname() {
			return true
		}
	}
	return false
}

// Reverses the bytes of a byte slice.
func ReverseBytes(b []byte) []byte {
	reversed := make([]byte, len(b))
	for i := range b {
		reversed[i] = b[len(b)-1-i]
	}
	return reversed
}

// Generates a short request ID.
func GenerateShortRequestID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "unknown-request-id"
	}
	return hex.EncodeToString(b) // e.g., "f4e3d2a1b3c0d9e8"
}

// SliceToU128 converts a byte slice to a big integer.
func SliceToU128(buf []byte) (*big.Int, error) {

	// Check if the buffer is 16 bytes.
	if len(buf) != 16 {
		return nil, errors.New("cannot convert slice to u128: invalid size")
	}

	// Create the result.
	result := big.NewInt(0)

	// Convert the buffer to a big integer.
	for idx, b := range buf {
		bigByte := big.NewInt(int64(b))
		bigByte.Lsh(bigByte, 8*uint(idx))
		result.Add(result, bigByte)
	}

	return result, nil
}

func GetRetryableHTTPClient(maxRetries int) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 3 * time.Second
	retryClient.RetryMax = maxRetries
	return retryClient
}

// getClientIP extracts the real client IP from various headers
func GetClientIP(r *http.Request) string {
	// Check for X-Real-IP header (set by nginx/proxy)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if isValidIP(realIP) {
			return realIP
		}
	}
	
	// Check for X-Forwarded-For header (set by load balancers)
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// Take the first valid IP (the original client)
		ips := strings.Split(fwd, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" && ip != "unknown" && isValidIP(ip) {
				return ip
			}
		}
	}
	
	// Check for X-Forwarded header
	if xForwarded := r.Header.Get("X-Forwarded"); xForwarded != "" {
		if isValidIP(xForwarded) {
			return xForwarded
		}
	}
	
	// Fallback to RemoteAddr (remove port if present)
	remoteAddr := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		remoteAddr = host
	}
	
	return remoteAddr
}

// isValidIP checks if a string is a valid IP address
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}