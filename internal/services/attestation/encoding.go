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

// prepareAttestationData prepares the attestation data.
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

// PrepareProofData prepares the proof data.
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

	// Add the padding to the URL length.
	// urlLenWithPadding := len(req.Url) + len(utils.GetPadding([]byte(req.Url), encoding.TARGET_ALIGNMENT))

	// Check if the URL length is too long.
	if urlLen > math.MaxUint16 {
		logger.Error("Cannot create encoded data meta header - urlLen is too long")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	selectorLen := len(req.Selector)

	// Add the padding to the selector length.
	// selectorLenWithPadding := len(req.Selector) + len(utils.GetPadding([]byte(req.Selector), encoding.TARGET_ALIGNMENT))

	// Check if the selector length is too long.
	if selectorLen > math.MaxUint16 {
		logger.Error("Cannot create encoded data meta header - selectorLen is too long")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	headersLen := len(encodedHeaders)

	// Check if the headers length is too long.
	if headersLen > math.MaxUint16 {
		logger.Error("Cannot create encoded data meta header - headersLen is too long")
		return nil, nil, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	optionalFieldsLen := len(encodedOptionalFields)

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

// PrepareEncodedRequestProof prepares the encoded request proof.
func PrepareEncodedRequestProof(userData []byte, encodedPositions encoding.ProofPositionalInfo) ([]byte, *appErrors.AppError) {

	// Get the attestation data length and timestamp length.
	attestationDataLen := encodedPositions.Data.Len
	timestampLen := encodedPositions.Timestamp.Len

	metaHeaderLen := 2 * encoding.TARGET_ALIGNMENT

	// Calculate the end offset.
	endOffset := metaHeaderLen + (attestationDataLen+timestampLen)*encoding.TARGET_ALIGNMENT

	// Check if the user data is too short.
	if endOffset > len(userData) {
		logger.Error("User data is too short")
		return nil, appErrors.NewAppError(appErrors.ErrUserDataTooShort)
	}

	clear(userData[metaHeaderLen:endOffset])

	return userData, nil
}
