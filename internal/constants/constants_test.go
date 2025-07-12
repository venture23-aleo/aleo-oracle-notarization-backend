package constants

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGRAMINE_PATHS_Structure(t *testing.T) {
	// Test that GRAMINE_PATHS is properly initialized
	expectedPaths := GraminePathsStruct{
		MY_TARGET_INFO_PATH:   "/dev/attestation/my_target_info",
		TARGET_INFO_PATH:      "/dev/attestation/target_info",
		USER_REPORT_DATA_PATH: "/dev/attestation/user_report_data",
		REPORT_PATH:           "/dev/attestation/report",
		ATTESTATION_TYPE_PATH: "/dev/attestation/attestation_type",
		QUOTE_PATH:            "/dev/attestation/quote",
	}

	assert.Equal(t, GRAMINE_PATHS, expectedPaths, "GRAMINE_PATHS should be properly initialized")
}

func TestGRAMINE_PATHS_IndividualFields(t *testing.T) {
	// Test individual fields
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{
			name:     "MY_TARGET_INFO_PATH",
			field:    GRAMINE_PATHS.MY_TARGET_INFO_PATH,
			expected: "/dev/attestation/my_target_info",
		},
		{
			name:     "TARGET_INFO_PATH",
			field:    GRAMINE_PATHS.TARGET_INFO_PATH,
			expected: "/dev/attestation/target_info",
		},
		{
			name:     "USER_REPORT_DATA_PATH",
			field:    GRAMINE_PATHS.USER_REPORT_DATA_PATH,
			expected: "/dev/attestation/user_report_data",
		},
		{
			name:     "REPORT_PATH",
			field:    GRAMINE_PATHS.REPORT_PATH,
			expected: "/dev/attestation/report",
		},
		{
			name:     "ATTESTATION_TYPE_PATH",
			field:    GRAMINE_PATHS.ATTESTATION_TYPE_PATH,
			expected: "/dev/attestation/attestation_type",
		},
		{
			name:     "QUOTE_PATH",
			field:    GRAMINE_PATHS.QUOTE_PATH,
			expected: "/dev/attestation/quote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.field, tt.expected, "%s should be properly initialized", tt.name)
		})
	}
}

func TestGRAMINE_PATHS_NotEmpty(t *testing.T) {
	// Test that all paths are not empty
	paths := []string{
		GRAMINE_PATHS.MY_TARGET_INFO_PATH,
		GRAMINE_PATHS.TARGET_INFO_PATH,
		GRAMINE_PATHS.USER_REPORT_DATA_PATH,
		GRAMINE_PATHS.REPORT_PATH,
		GRAMINE_PATHS.ATTESTATION_TYPE_PATH,
		GRAMINE_PATHS.QUOTE_PATH,
	}

	for i, path := range paths {
		assert.NotEmpty(t, path, "Path at index %d should not be empty", i)
	}
}

func TestGRAMINE_PATHS_StartWithDevAttestation(t *testing.T) {
	// Test that all paths start with "/dev/attestation/"
	paths := []string{
		GRAMINE_PATHS.MY_TARGET_INFO_PATH,
		GRAMINE_PATHS.TARGET_INFO_PATH,
		GRAMINE_PATHS.USER_REPORT_DATA_PATH,
		GRAMINE_PATHS.REPORT_PATH,
		GRAMINE_PATHS.ATTESTATION_TYPE_PATH,
		GRAMINE_PATHS.QUOTE_PATH,
	}

	expectedPrefix := "/dev/attestation/"
	for i, path := range paths {
		assert.True(t, len(path) >= len(expectedPrefix) && path[:len(expectedPrefix)] == expectedPrefix, "Path at index %d (%s) should start with %s", i, path, expectedPrefix)
	}
}

func TestALLOWED_HEADERS_NotEmpty(t *testing.T) {
	// Test that ALLOWED_HEADERS is not empty
	assert.NotEmpty(t, ALLOWED_HEADERS, "ALLOWED_HEADERS should not be empty")
}

func TestALLOWED_HEADERS_ContainsExpectedHeaders(t *testing.T) {
	// Test that ALLOWED_HEADERS contains expected common headers
	expectedHeaders := []string{
		"Accept",
		"Content-Type",
		"User-Agent",
		"Host",
		"Cache-Control",
		"Connection",
		"Date",
		"Origin",
		"Referer",
	}

	for _, expectedHeader := range expectedHeaders {
		assert.Contains(t, ALLOWED_HEADERS, expectedHeader, "ALLOWED_HEADERS should contain expected header: %s", expectedHeader)
	}
}

func TestALLOWED_HEADERS_NoDuplicates(t *testing.T) {
	// Test that ALLOWED_HEADERS contains no duplicates
	headerMap := make(map[string]bool)
	for _, header := range ALLOWED_HEADERS {
		assert.False(t, headerMap[header], "ALLOWED_HEADERS contains duplicate header: %s", header)
		headerMap[header] = true
	}
}

func TestALLOWED_HEADERS_AllNotEmpty(t *testing.T) {
	// Test that all headers in ALLOWED_HEADERS are not empty
	assert.NotEmpty(t, ALLOWED_HEADERS, "ALLOWED_HEADERS should not be empty")
}

func TestALLOWED_HEADERS_ContainsSecurityHeaders(t *testing.T) {
	// Test that ALLOWED_HEADERS contains important security-related headers
	securityHeaders := []string{
		"X-Forwarded-For",
		"X-Forwarded-Host",
		"X-Forwarded-Proto",
		"X-Requested-With",
	}

	for _, securityHeader := range securityHeaders {
		assert.Contains(t, ALLOWED_HEADERS, securityHeader, "ALLOWED_HEADERS should contain security header: %s", securityHeader)
	}
}

func TestGraminePathsStruct_FieldNames(t *testing.T) {
	// Test that the struct has the expected field names
	pathsType := reflect.TypeOf(GRAMINE_PATHS)
	expectedFields := []string{
		"MY_TARGET_INFO_PATH",
		"TARGET_INFO_PATH",
		"USER_REPORT_DATA_PATH",
		"REPORT_PATH",
		"ATTESTATION_TYPE_PATH",
		"QUOTE_PATH",
	}

	for _, expectedField := range expectedFields {
		_, found := pathsType.FieldByName(expectedField)
		assert.True(t, found, "GraminePathsStruct missing expected field: %s", expectedField)
	}
}
