package attestation

import (
	"bytes"
	"math"
	"strings"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	"github.com/venture23-aleo/aleo-oracle-encoding/positionRecorder"
)

// prepareAttestationData formats and pads the attestation data string according to the specified encoding option.
//
// This function takes the raw attestation data and the encoding options, then processes the data based on the encoding type:
//   - For string encoding, it pads the string to the ATTESTATION_DATA_SIZE_LIMIT using null bytes.
//   - For float encoding, it ensures the string contains a decimal point and pads with '0' to MaxUint8 length.
//   - For integer encoding, it prepends '0' characters to the string to reach MaxUint8 length, allowing for consistent parsing.
//
// If an invalid encoding option is provided, it returns an error.
//
// Parameters:
//   - attestationData: The raw attestation data as a string.
//   - encodingOptions: Pointer to the encoding options specifying the data type.
//
// Returns:
//   - string: The padded and formatted attestation data.
//   - *appErrors.AppError: An error if the encoding option is invalid or padding fails.
func prepareAttestationData(attestationData string, encodingOptions *encoding.EncodingOptions) (string, *appErrors.AppError) {

	// Check the encoding option.
	switch encodingOptions.Value {
	case encoding.ENCODING_OPTION_STRING:
		return utils.PadStringToLength(attestationData, 0x00, constants.ATTESTATION_DATA_SIZE_LIMIT)

	case encoding.ENCODING_OPTION_FLOAT:
		// Check if the attestation data contains a dot.
		if strings.Contains(attestationData, ".") {
			return utils.PadStringToLength(attestationData, '0', math.MaxUint8)
		} else {
			// Pad the string to the target length.
			return utils.PadStringToLength(attestationData+".", '0', math.MaxUint8)
		}
	case encoding.ENCODING_OPTION_INT:
		// For integers we prepend zeroes instead of appending, that allows strconv to parse it no matter how many zeroes there are
		padString, err := utils.PadStringToLength("", '0', math.MaxUint8-len(attestationData))
		if err != nil {
			return "", err
		}
		return padString + attestationData, nil
	default:
		return "", appErrors.NewAppError(appErrors.ErrInvalidEncodingOption)
	}
}

// PrepareProofData encodes all attestation request fields and metadata into a single aligned byte buffer for proof generation,
// returning the encoded buffer, positional information for each field, and any error encountered during the process.
func PrepareProofData(statusCode int, attestationData string, timestamp int64, req AttestationRequest) ([]byte, *encoding.ProofPositionalInfo, *appErrors.AppError) {

	// Prepare the attestation data.
	var preppedAttestationData string

	// Check if the URL is a price feed.
	if req.Url != constants.PRICE_FEED_BTC_URL && req.Url != constants.PRICE_FEED_ETH_URL && req.Url != constants.PRICE_FEED_ALEO_URL {
		var preppedAttestationDataError *appErrors.AppError
		preppedAttestationData, preppedAttestationDataError = prepareAttestationData(attestationData, &req.EncodingOptions)
		if preppedAttestationDataError != nil {
			logger.Error("Failed to prepare attestation data: ", "error", preppedAttestationDataError)
			return nil, nil, preppedAttestationDataError
		}
	} else {
		preppedAttestationData = attestationData
	}

	// Create the buffer.
	var buf bytes.Buffer

	// Create the position recorder.
	recorder := positionRecorder.NewPositionRecorder(&buf, encoding.TARGET_ALIGNMENT)

	// Write an empty meta header.
	encoding.WriteWithPadding(recorder, make([]byte, encoding.TARGET_ALIGNMENT*2))

	// Write the attestation data.
	attestationDataBuffer, err := encoding.EncodeAttestationData(preppedAttestationData, &req.EncodingOptions)

	// Check if the attestation data is encoded.
	if err != nil {
		logger.Error("Failed to encode attestation data: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrEncodingAttestationData)
	}

	// Write the attestation data to the buffer.
	attesatationDataPositionInfo, err := encoding.WriteWithPadding(recorder, attestationDataBuffer)

	if err != nil {
		logger.Error("Failed to write attestation data to buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingAttestationData)
	}

	// Write the timestamp to the buffer.
	timestampPositionInfo, err := encoding.WriteWithPadding(recorder, encoding.NumberToBytes(uint64(timestamp)))

	if err != nil {
		logger.Error("Failed to write timestamp to buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingTimestamp)
	}

	// Write the status code to the buffer.
	statusCodePositionInfo, err := encoding.WriteWithPadding(recorder, encoding.NumberToBytes(uint64(statusCode)))

	if err != nil {
		logger.Error("Failed to write status code to buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingStatusCode)
	}

	// Write the URL to the buffer.
	urlPositionInfo, err := encoding.WriteWithPadding(recorder, []byte(req.Url))

	if err != nil {
		logger.Error("Failed to write URL to buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingUrl)
	}

	// Write the selector to the buffer.
	selectorPositionInfo, err := encoding.WriteWithPadding(recorder, []byte(req.Selector))

	if err != nil {
		logger.Error("Failed to write selector to buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingSelector)
	}

	// Encode the response format.
	responseFormat, err := encoding.EncodeResponseFormat(req.ResponseFormat)
	if err != nil {
		logger.Error("Failed to encode response format: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrEncodingResponseFormat)
	}

	// Write the response format to the buffer.
	responseFormatPositionInfo, err := encoding.WriteWithPadding(recorder, responseFormat)

	if err != nil {
		logger.Error("Failed to write response format to buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingResponseFormat)
	}

	// Write the request method to the buffer.
	requestMethodPositionInfo, err := encoding.WriteWithPadding(recorder, []byte(req.RequestMethod))

	if err != nil {
		logger.Error("Failed to write request method to buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingRequestMethod)
	}

	// Encode the encoding options.
	encodingOptions, err := encoding.EncodeEncodingOptions(&req.EncodingOptions)
	if err != nil {
		logger.Error("Failed to encode encoding options: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrEncodingEncodingOptions)
	}

	// Write the encoding options to the buffer.
	encodingOptionsPositionInfo, err := encoding.WriteWithPadding(recorder, encodingOptions)

	if err != nil {
		logger.Error("Failed to write encoding options to buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingEncodingOptions)
	}

	// Encode the request headers.
	encodedHeaders := encoding.EncodeHeaders(req.RequestHeaders)

	// Write the request headers to the buffer.
	requestHeadersPositionInfo, err := encoding.WriteWithPadding(recorder, encodedHeaders)

	if err != nil {
		logger.Error("Failed to write request headers to buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingRequestHeaders)
	}

	// write optional fields:
	// - html result type (exists only if response format is html)
	// - request content-type (can exist only if method is POST)
	// - request body (can exist only if method is POST)

	// Encode the optional fields.
	encodedOptionalFields, err := encoding.EncodeOptionalFields(req.HTMLResultType, req.RequestContentType, req.RequestBody)
	if err != nil {
		logger.Error("Failed to encode optional fields: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrEncodingOptionalFields)
	}

	// Write the optional fields to the buffer.
	optionalFieldsPositionInfo, err := encoding.WriteWithPadding(recorder, encodedOptionalFields)

	if err != nil {
		logger.Error("Failed to write optional fields buffer: ", "error", err)
		return nil, nil, appErrors.NewAppError(appErrors.ErrWrittingOptionalFields)
	}

	result := buf.Bytes()

	// Check if the result is aligned.
	if len(result)%encoding.TARGET_ALIGNMENT != 0 {
		logger.Error("Result is not aligned!")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	logger.Debug("preppedAttestationData: ", "preppedAttestationData", preppedAttestationData)

	attestationDataLen := len(preppedAttestationData)

	// Check if the attestation data length is too long.
	if attestationDataLen > math.MaxUint16 {
		logger.Error("Cannot create encoded data meta header - attestationDataLen is too long")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	methodLen := len(req.RequestMethod)
	logger.Debug("methodLen", "methodLen", methodLen)

	// Check if the method length is too long.
	if methodLen > math.MaxUint16 {
		logger.Error("Cannot create encoded data meta header - methodLen is too long")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	urlLen := len(req.Url)
	logger.Debug("urlLen", "urlLen", urlLen)

	// Check if the URL length is too long.
	if urlLen > math.MaxUint16 {
		logger.Error("Cannot create encoded data meta header - urlLen is too long")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	selectorLen := len(req.Selector)
	logger.Debug("selectorLen", "selectorLen", selectorLen)

	// Check if the selector length is too long.
	if selectorLen > math.MaxUint16 {
		logger.Error("Cannot create encoded data meta header - selectorLen is too long")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	headersLen := len(encodedHeaders)
	logger.Debug("headersLen", "headersLen", headersLen)

	// Check if the headers length is too long.
	if headersLen > math.MaxUint16 {
		logger.Error("Cannot create encoded data meta header - headersLen is too long")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	optionalFieldsLen := len(encodedOptionalFields)
	logger.Debug("optionalFieldsLen", "optionalFieldsLen", optionalFieldsLen)

	// Check if the optional fields length is too long.
	if optionalFieldsLen > math.MaxUint16 {
		logger.Error("Cannot create encoded data meta header - optionalFieldsLen is too long")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	// Create the meta header. Fill the empty meta header with the actual content.
	encoding.CreateMetaHeader(
		result[:encoding.TARGET_ALIGNMENT*2],
		uint16(attestationDataLen),
		uint16(methodLen),
		uint16(urlLen),
		uint16(selectorLen),
		uint16(headersLen),
		uint16(optionalFieldsLen),
	)

	proofPositionalInfo := &encoding.ProofPositionalInfo{
		Data:            *attesatationDataPositionInfo,
		Timestamp:       *timestampPositionInfo,
		StatusCode:      *statusCodePositionInfo,
		Method:          *requestMethodPositionInfo,
		ResponseFormat:  *responseFormatPositionInfo,
		Url:             *urlPositionInfo,
		Selector:        *selectorPositionInfo,
		EncodingOptions: *encodingOptionsPositionInfo,
		RequestHeaders:  *requestHeadersPositionInfo,
		OptionalFields:  *optionalFieldsPositionInfo,
	}

	return result, proofPositionalInfo, nil
}

// PrepareEncodedRequestProof zeroes out the attestation data and timestamp fields in the userData buffer
// based on their encoded positions, producing a canonicalized request encoding for proof purposes.
//
// This function is used to canonicalize the userData buffer for proof generation by zeroing out the
// attestation data and timestamp fields. This ensures that the resulting encoded request is deterministic
// and does not leak any sensitive or variable information in these fields.
//
// The function performs the following steps sequentially:
//  1. Retrieves the lengths of the attestation data and timestamp fields from the encodedPositions struct.
//  2. Calculates the length of the meta header (always 2 * encoding.TARGET_ALIGNMENT).
//  3. Computes the end offset in the userData buffer that covers both the attestation data and timestamp fields.
//  4. Checks if the userData buffer is large enough to accommodate the zeroing operation.
//  5. Zeroes out the relevant section of the userData buffer using the built-in clear function.
//  6. Returns the modified userData buffer and nil error on success, or an error if the buffer is too short.
//
// Parameters:
//   - userData ([]byte): The original user data buffer to be canonicalized. This buffer will be modified in place.
//   - encodedPositions (encoding.ProofPositionalInfo): Struct containing the encoded positions and lengths
//     of the attestation data and timestamp fields within the userData buffer.
//
// Returns:
//   - ([]byte): The canonicalized userData buffer with attestation data and timestamp fields zeroed out.
//   - (*appErrors.AppError): An application error if the operation fails (e.g., if the buffer is too short).
func PrepareEncodedRequestProof(userData []byte, encodedPositions encoding.ProofPositionalInfo) ([]byte, *appErrors.AppError) {
	// Step 1: Retrieve the attestation data and timestamp lengths from the encoded positions.
	attestationDataLen := encodedPositions.Data.Len
	timestampLen := encodedPositions.Timestamp.Len

	logger.Debug("attestationDataLen", "attestationDataLen", attestationDataLen)
	logger.Debug("timestampLen", "timestampLen", timestampLen)

	// Step 2: Calculate the meta header length (fixed size).
	metaHeaderLen := 2 * encoding.TARGET_ALIGNMENT

	// Step 3: Calculate the end offset for the section to be zeroed.
	endOffset := metaHeaderLen + (attestationDataLen+timestampLen)*encoding.TARGET_ALIGNMENT

	logger.Debug("endOffset", "endOffset", endOffset)

	// Step 4: Check if the userData buffer is large enough.
	if endOffset > len(userData) {
		logger.Error("User data is too short")
		return nil, appErrors.NewAppError(appErrors.ErrUserDataTooShort)
	}

	// Log the userData buffer before zeroing.
	logger.Debug("userData before clearing", "userData", userData)

	// Step 5: Zero out the attestation data and timestamp fields in the buffer.
	clear(userData[metaHeaderLen:endOffset])

	// Log the userData buffer after zeroing.
	logger.Debug("userData after clearing", "userData", userData)

	// Return the canonicalized buffer and nil error.
	return userData, nil
}
