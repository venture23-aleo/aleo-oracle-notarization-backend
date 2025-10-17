// Package sgx implements SGX quote generation and Open Enclave wrapping helpers.
package sgx

import (
	"bytes"
	"encoding/binary"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

// wrapRawQuoteAsOpenEnclaveEvidence wraps a raw SGX quote as Open Enclave evidence format.
//
// This function constructs an Open Enclave evidence buffer by prepending the required headers
// (version, type, and quote length) to the raw quote buffer. The resulting byte slice is formatted as follows:
//   - 4 bytes: Open Enclave version (little-endian uint32, currently 1)
//   - 4 bytes: Open Enclave type (little-endian uint32, currently 2)
//   - 8 bytes: Quote length (little-endian uint32, padded to 8 bytes)
//   - N bytes: Raw quote buffer
//
// This format is required by the verifier backend and the contract, which expects the quote to be wrapped as Open Enclave evidence.
//
// Parameters:
//   - rawQuoteBuffer ([]byte): The raw SGX quote to be wrapped.
//
// Returns:
//   - []byte: The Open Enclave evidence buffer containing the headers and the raw quote.
func wrapRawQuoteAsOpenEnclaveEvidence(rawQuoteBuffer []byte) []byte {
	const (
		oeVersionConst = 1 // Open Enclave evidence version
		oeVersionLen   = 4 // Length of version field in bytes
		oeTypeConst    = 2 // Open Enclave evidence type (2 = SGX ECDSA)
		oeTypeLen      = 4 // Length of type field in bytes
		oeQuoteLen     = 8 // Length of quote length field in bytes
	)

	// Create the Open Enclave version header (4 bytes, little-endian)
	oeVersion := make([]byte, oeVersionLen)
	binary.LittleEndian.PutUint32(oeVersion, oeVersionConst)

	// Create the Open Enclave type header (4 bytes, little-endian)
	oeType := make([]byte, oeTypeLen)
	binary.LittleEndian.PutUint32(oeType, oeTypeConst)

	// Create the quote length header (8 bytes, little-endian, only lower 4 bytes used)
	quoteLength := make([]byte, oeQuoteLen)
	binary.LittleEndian.PutUint32(quoteLength, uint32(len(rawQuoteBuffer)))

	// Assemble the Open Enclave evidence buffer
	var buf bytes.Buffer
	buf.Write(oeVersion)      // Write version header
	buf.Write(oeType)         // Write type header
	buf.Write(quoteLength)    // Write quote length header
	buf.Write(rawQuoteBuffer) // Write the raw quote

	// Return the complete Open Enclave evidence as a byte slice
	return buf.Bytes()
}

// GenerateQuote generates a quote for the attestation service.
// Refer to https://gramine.readthedocs.io/en/stable/attestation.html#low-level-dev-attestation-interface
//
// This function performs the following steps sequentially:
// 1. Acquires a lock to ensure thread-safe quote generation (as quote generation is not thread-safe).
// 2. Prepares a 64-byte report data buffer, copying the provided inputData into it (truncating or zero-padding as needed).
// 3. Writes the report data to the user report data path specified in the Gramine configuration.
// 4. Reads the raw SGX quote from the Gramine quote path.
// 5. Wraps the raw quote as Open Enclave evidence, as required by the contract and verifier backend.
// 6. Returns the wrapped quote as a byte slice, or an error if any step fails.
//
// Parameters:
//   - inputData ([]byte): The input data to be included in the report data (will be truncated or zero-padded to 64 bytes).
//
// Returns:
//   - []byte: The Open Enclave evidence buffer containing the wrapped quote.
//   - *appErrors.AppError: An application error if any step fails, otherwise nil.
func GenerateQuote(inputData []byte) ([]byte, *appErrors.AppError) {
	// Step 1: Acquire the enclave lock for thread safety.
	enclaveLock.Lock()
	defer enclaveLock.Unlock()

	// Step 2: Prepare the 64-byte report data buffer.
	if len(inputData) > 64 {
		logger.Error("input data too large for SGX report data", "max", 64, "got", len(inputData))
		return nil, appErrors.ErrInvalidSGXReportSize
	}
	reportData := make([]byte, 64)
	copy(reportData, inputData) // Copy inputData (truncates or zero-pads as needed)

	// Step 3: Write the report data to the user report data path.
	err := SecureWriteFile(GraminePseudoFilesRoot, gramineAttestationPaths.UserReportDataPath, reportData)
	if err != nil {
		logger.Error("Error while writing report data:", "error", err)
		return nil, appErrors.ErrWritingReportData
	}

	// Step 4: Read the raw quote from the Gramine quote path.
	quote, err := SecureReadFile(GraminePseudoFilesRoot, gramineAttestationPaths.QuotePath)
	if err != nil {
		logger.Error("Error while reading quote: ", "error", err)
		return nil, appErrors.ErrReadingQuote
	}

	if len(quote) < QuoteMinSize {
		logger.Error("Quote is too small", "quote", len(quote))
		return nil, appErrors.ErrInvalidSGXQuoteSize
	}

	// Step 5: Wrap the raw quote as Open Enclave evidence.
	finalQuote := wrapRawQuoteAsOpenEnclaveEvidence(quote)

	// Step 6: Return the final wrapped quote.
	return finalQuote, nil
}
