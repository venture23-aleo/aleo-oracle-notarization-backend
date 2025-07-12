package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
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
