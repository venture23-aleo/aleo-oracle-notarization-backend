package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
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
func PadStringToLength(str string, paddingChar byte, targetLength int) (string, *appErrors.AppError) {
	if len(str) > targetLength {
		logger.Error("PadStringToLength: string length is greater than target length", "string length", len(str), "target length", targetLength)
		return "", appErrors.NewAppError(appErrors.ErrAttestationDataTooLarge)
	}
	// Pad the string to the target length.
	return str + strings.Repeat(string(paddingChar), targetLength-len(str)), nil
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
