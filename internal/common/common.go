package common

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"net/url"
	"strings"

	"github.com/cloudflare/roughtime/client"
	configs "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/config"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

// IsPriceFeedURL checks if the URL is a price feed URL.
func IsPriceFeedURL(url string) bool {
	return url == constants.PriceFeedBTCURL ||
		url == constants.PriceFeedETHURL ||
		url == constants.PriceFeedAleoURL
}

// ExtractAssetFromPriceFeedURL extracts the asset name from price feed URL
func ExtractTokenFromPriceFeedURL(url string) string {
	switch url {
	case constants.PriceFeedBTCURL:
		return "BTC"
	case constants.PriceFeedETHURL:
		return "ETH"
	case constants.PriceFeedAleoURL:
		return "ALEO"
	default:
		return "UNKNOWN"
	}
}

// GetTokenIDFromPriceFeedURL gets the token ID from price feed URL
func GetTokenIDFromPriceFeedURL(url string) int {
	switch url {
	case constants.PriceFeedBTCURL:
		return constants.BTCTokenID
	case constants.PriceFeedETHURL:
		return constants.ETHTokenID
	case constants.PriceFeedAleoURL:
		return constants.AleoTokenID
	default:
		return 0
	}
}

// NormalizeURL adds https:// if the scheme is missing and validates the result.
func NormalizeURL(rawURL string) (string, *appErrors.AppError) {
	if !strings.HasPrefix(strings.ToLower(rawURL), "http://") && !strings.HasPrefix(strings.ToLower(rawURL), "https://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.ParseRequestURI(rawURL)

	if err != nil || !strings.Contains(parsed.Hostname(), ".") {
		return "", appErrors.ErrInvalidURL
	}

	return parsed.String(), nil
}

func GetHostnameFromURL(rawURL string) (string, *appErrors.AppError) {
	if !strings.HasPrefix(strings.ToLower(rawURL), "http://") && !strings.HasPrefix(strings.ToLower(rawURL), "https://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || !strings.Contains(parsed.Hostname(), ".") {
		return "", appErrors.ErrInvalidURL
	}

	return parsed.Hostname(), nil
}

// IsAcceptedHeader checks if a header name is in the list of allowed headers.
func IsAcceptedHeader(header string) bool {
	for _, h := range constants.AllowedHeaders {
		if strings.ToLower(strings.TrimSpace(h)) == header {
			return true
		}
	}
	return false
}

// IsTargetWhitelisted checks if a target is in the list of whitelisted domains.
func IsTargetWhitelisted(endpoint string) bool {
	if IsPriceFeedURL(endpoint) {
		return true
	}

	hostname, err := GetHostnameFromURL(endpoint)
	if err != nil {
		return false
	}

	for _, domainName := range configs.GetWhitelistedDomains() {
		if strings.ToLower(strings.TrimSpace(domainName)) == hostname {
			return true
		}
	}

	return false
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

// padStringToLength pads the string to the target length.
func PadStringToLength(str string, paddingChar byte, targetLength int) (string, *appErrors.AppError) {
	if len(str) > targetLength {
		logger.Error("PadStringToLength: string length is greater than target length", "string length", len(str), "target length", targetLength)
		return "", appErrors.ErrAttestationDataTooLarge
	}
	// Pad the string to the target length.
	return str + strings.Repeat(string(paddingChar), targetLength-len(str)), nil
}

func SliceToU128(buf []byte) (*big.Int, *appErrors.AppError) {
	if len(buf) != 16 {
		return nil, appErrors.ErrSliceToU128
	}

	result := big.NewInt(0)

	for idx, b := range buf {
		bigByte := big.NewInt(int64(b))
		bigByte.Lsh(bigByte, 8*uint(idx))
		result.Add(result, bigByte)
	}

	return result, nil
}

func GetTimestampFromRoughtime() (int64, *appErrors.AppError) { 
	roughtimeConfig := configs.GetRoughtimeConfig()

	server := roughtimeConfig.ServerConfig.Server

	// Query roughtime servers with timeout and retry settings
	rt,err := client.Get(server, roughtimeConfig.Retries, roughtimeConfig.Timeout, nil)

	if err != nil {
		logger.Error("Roughtime query failed: %s\n", err)
		return 0, appErrors.ErrRoughtimeServerError
	}

	t := rt.Midpoint.UTC()

	return t.Unix(), nil
}