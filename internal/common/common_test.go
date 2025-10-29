package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

func TestIsPriceFeedURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "BTC price feed URL",
			url:      constants.PriceFeedBTCURL,
			expected: true,
		},
		{
			name:     "ETH price feed URL",
			url:      constants.PriceFeedETHURL,
			expected: true,
		},
		{
			name:     "ALEO price feed URL",
			url:      constants.PriceFeedAleoURL,
			expected: true,
		},
		{
			name:     "Regular URL",
			url:      "https://example.com",
			expected: false,
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: false,
		},
		{
			name:     "Similar but different URL",
			url:      "price_feed: btc_other",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPriceFeedURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTokenFromPriceFeedURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "BTC price feed URL",
			url:      constants.PriceFeedBTCURL,
			expected: "BTC",
		},
		{
			name:     "ETH price feed URL",
			url:      constants.PriceFeedETHURL,
			expected: "ETH",
		},
		{
			name:     "ALEO price feed URL",
			url:      constants.PriceFeedAleoURL,
			expected: "ALEO",
		},
		{
			name:     "Unknown URL",
			url:      "https://example.com",
			expected: "UNKNOWN",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "UNKNOWN",
		},
		{
			name:     "Similar but different URL",
			url:      "price_feed: btc_other",
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTokenFromPriceFeedURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTokenIDFromPriceFeedURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{
			name:     "BTC price feed URL",
			url:      constants.PriceFeedBTCURL,
			expected: constants.BTCTokenID,
		},
		{
			name:     "ETH price feed URL",
			url:      constants.PriceFeedETHURL,
			expected: constants.ETHTokenID,
		},
		{
			name:     "ALEO price feed URL",
			url:      constants.PriceFeedAleoURL,
			expected: constants.AleoTokenID,
		},
		{
			name:     "Unknown URL",
			url:      "https://example.com",
			expected: 0,
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: 0,
		},
		{
			name:     "Similar but different URL",
			url:      "price_feed: btc_other",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTokenIDFromPriceFeedURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "URL with https scheme",
			input:       "https://example.com",
			expected:    "https://example.com",
			expectError: false,
		},
		{
			name:        "URL with http scheme",
			input:       "http://example.com",
			expected:    "http://example.com",
			expectError: false,
		},
		{
			name:        "URL without scheme",
			input:       "example.com",
			expected:    "https://example.com",
			expectError: false,
		},
		{
			name:        "URL with path",
			input:       "example.com/path",
			expected:    "https://example.com/path",
			expectError: false,
		},
		{
			name:        "URL with query parameters",
			input:       "example.com?param=value",
			expected:    "https://example.com?param=value",
			expectError: false,
		},
		{
			name:        "URL with fragment - should fail",
			input:       "example.com#fragment",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid URL - no domain",
			input:       "not-a-url",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid URL - empty",
			input:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid URL - just scheme",
			input:       "https://",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid URL - malformed",
			input:       "://example.com",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid URL - no dot in hostname",
			input:       "localhost",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeURL(tt.input)
			if tt.expectError {
				assert.NotNil(t, err)
				assert.Equal(t, appErrors.ErrInvalidURL, err)
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetHostnameFromURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "URL with https scheme",
			input:       "https://example.com",
			expected:    "example.com",
			expectError: false,
		},
		{
			name:        "URL with http scheme",
			input:       "http://example.com",
			expected:    "example.com",
			expectError: false,
		},
		{
			name:        "URL without scheme",
			input:       "example.com",
			expected:    "example.com",
			expectError: false,
		},
		{
			name:        "URL with subdomain",
			input:       "sub.example.com",
			expected:    "sub.example.com",
			expectError: false,
		},
		{
			name:        "URL with port",
			input:       "example.com:8080",
			expected:    "example.com",
			expectError: false,
		},
		{
			name:        "URL with path",
			input:       "example.com/path",
			expected:    "example.com",
			expectError: false,
		},
		{
			name:        "URL with query parameters",
			input:       "example.com?param=value",
			expected:    "example.com",
			expectError: false,
		},
		{
			name:        "Invalid URL - no domain",
			input:       "not-a-url",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid URL - empty",
			input:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid URL - just scheme",
			input:       "https://",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid URL - no dot in hostname",
			input:       "localhost",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetHostnameFromURL(tt.input)
			if tt.expectError {
				assert.NotNil(t, err)
				assert.Equal(t, appErrors.ErrInvalidURL, err)
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsAcceptedHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{
			name:     "Valid header - Accept",
			header:   "Accept",
			expected: true,
		},
		{
			name:     "Valid header - Content-Type",
			header:   "Content-Type",
			expected: true,
		},
		{
			name:     "Valid header - User-Agent",
			header:   "User-Agent",
			expected: true,
		},
		{
			name:     "Valid header with spaces",
			header:   "Accept",
			expected: true,
		},
		{
			name:     "Invalid header",
			header:   "X-Custom-Header",
			expected: false,
		},
		{
			name:     "Empty header",
			header:   "",
			expected: false,
		},
		{
			name:     "Case sensitive - should match exactly",
			header:   "accept",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAcceptedHeader(tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTargetWhitelisted(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected bool
	}{
		{
			name:     "BTC price feed URL",
			endpoint: constants.PriceFeedBTCURL,
			expected: true,
		},
		{
			name:     "ETH price feed URL",
			endpoint: constants.PriceFeedETHURL,
			expected: true,
		},
		{
			name:     "ALEO price feed URL",
			endpoint: constants.PriceFeedAleoURL,
			expected: true,
		},
		{
			name:     "Invalid URL",
			endpoint: "not-a-url",
			expected: false,
		},
		{
			name:     "Empty URL",
			endpoint: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTargetWhitelisted(tt.endpoint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateShortRequestID(t *testing.T) {
	// Test multiple generations to ensure uniqueness
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id := GenerateShortRequestID()

		// Check that ID is not empty
		assert.NotEmpty(t, id)

		// Check that ID is 32 characters long (16 bytes = 32 hex chars)
		assert.Equal(t, 32, len(id))

		// Check that ID contains only hex characters
		assert.True(t, isHexString(id), "ID should contain only hex characters")

		// Check that ID is unique
		assert.False(t, ids[id], "Generated ID should be unique")
		ids[id] = true
	}
}

func TestPadStringToLength(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		paddingChar  byte
		targetLength int
		expected     string
		expectError  bool
	}{
		{
			name:         "Pad short string",
			input:        "test",
			paddingChar:  '0',
			targetLength: 8,
			expected:     "test0000",
			expectError:  false,
		},
		{
			name:         "String already at target length",
			input:        "test",
			paddingChar:  '0',
			targetLength: 4,
			expected:     "test",
			expectError:  false,
		},
		{
			name:         "Empty string padding",
			input:        "",
			paddingChar:  '0',
			targetLength: 5,
			expected:     "00000",
			expectError:  false,
		},
		{
			name:         "String longer than target",
			input:        "very-long-string",
			paddingChar:  '0',
			targetLength: 5,
			expected:     "",
			expectError:  true,
		},
		{
			name:         "Zero target length - should error",
			input:        "test",
			paddingChar:  '0',
			targetLength: 0,
			expected:     "",
			expectError:  true,
		},
		{
			name:         "Negative target length - should error",
			input:        "test",
			paddingChar:  '0',
			targetLength: -1,
			expected:     "",
			expectError:  true,
		},
		{
			name:         "Different padding character",
			input:        "test",
			paddingChar:  '*',
			targetLength: 8,
			expected:     "test****",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PadStringToLength(tt.input, tt.paddingChar, tt.targetLength)
			if tt.expectError {
				assert.NotNil(t, err)
				assert.Equal(t, appErrors.ErrAttestationDataTooLarge, err)
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.targetLength, len(result))
			}
		})
	}
}

// Helper function to check if a string contains only hex characters
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}