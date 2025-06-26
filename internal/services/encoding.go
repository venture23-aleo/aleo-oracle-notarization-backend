package services

import (
	"bytes"
	"log"
	"math"
	"strings"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	encoding "github.com/zkportal/aleo-oracle-encoding"
	"github.com/zkportal/aleo-oracle-encoding/positionRecorder"
)

// prepareAttestationData prepares the attestation data.
func prepareAttestationData(attestationData string, encodingOptions *encoding.EncodingOptions) string {

	// Check the encoding option.
	switch encodingOptions.Value {
	case encoding.ENCODING_OPTION_STRING:
		// Pad the string to the target length.
		return utils.PadStringToLength(attestationData, 0x00, constants.AttestationDataSizeLimit)
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
		return utils.PadStringToLength("", '0', math.MaxUint8-len(attestationData)) + attestationData
	}

	return attestationData
}


// PrepareProofData prepares the proof data.
func PrepareProofData(statusCode int, attestationData string, timestamp int64, req AttestationRequest) ([]byte, ProofPositionalInfo, *appErrors.AppError) {

	// Prepare the attestation data.
	preppedAttestationData := attestationData

	// Check if the URL is a price feed.
	if req.Url != constants.PriceFeedBtcUrl && req.Url != constants.PriceFeedEthUrl && req.Url != constants.PriceFeedAleoUrl {
		preppedAttestationData = prepareAttestationData(attestationData, &req.EncodingOptions)
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
		log.Println("prepareProofData: failed to encode attestation data, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrEncodingAttestationData)
	}

	// Write the attestation data to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, attestationDataBuffer); err != nil {
		log.Println("prepareProofData: failed to write attestation data to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingAttestationData)
	}

	// Write the timestamp to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encoding.NumberToBytes(uint64(timestamp))); err != nil {
		log.Println("prepareProofData: failed to write timestamp to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingTimestamp)
	}

	// Write the status code to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encoding.NumberToBytes(uint64(statusCode))); err != nil {
		log.Println("prepareProofData: failed to write status code to buffer, err = ", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingStatusCode)
	}

	// Write the URL to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, []byte(req.Url)); err != nil {
		log.Println("prepareProofData: failed to write URL to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingUrl)
	}

	// Write the selector to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, []byte(req.Selector)); err != nil {
		log.Println("prepareProofData: failed to write selector to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingSelector)
	}

	// Encode the response format.
	responseFormat, err := encoding.EncodeResponseFormat(req.ResponseFormat)
	if err != nil {
		log.Println("prepareProofData: failed to encode response format, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrEncodingResponseFormat)
	}

	// Write the response format to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, responseFormat); err != nil {
		log.Println("prepareProofData: failed to write response format to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingResponseFormat)
	}

	// Write the request method to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, []byte(req.RequestMethod)); err != nil {
		log.Println("prepareProofData: failed to write request method to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingRequestMethod)
	}

	// Encode the encoding options.
	encodingOptions, err := encoding.EncodeEncodingOptions(&req.EncodingOptions)
	if err != nil {
		log.Println("prepareProofData: failed to encode encoding options, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrEncodingEncodingOptions)
	}

	// Write the encoding options to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encodingOptions); err != nil {
		log.Println("prepareProofData: failed to write encoding options to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingEncodingOptions)
	}

	// Encode the request headers.
	encodedHeaders := encoding.EncodeHeaders(req.RequestHeaders)

	// Write the request headers to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encodedHeaders); err != nil {
		log.Println("prepareProofData: failed to write request headers to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingRequestHeaders)
	}

	// write optional fields:
	// - html result type (exists only if response format is html)
	// - request content-type (can exist only if method is POST)
	// - request body (can exist only if method is POST)

	// Encode the optional fields.
	encodedOptionalFields, err := encoding.EncodeOptionalFields(req.HTMLResultType, req.RequestContentType, req.RequestBody)
	if err != nil {
		log.Println("prepareProofData: failed to encode optional fields, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrEncodingOptionalFields)
	}

	// Write the optional fields to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encodedOptionalFields); err != nil {
		log.Println("prepareProofData: failed to write optional fields buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrWrittingOptionalFields)
	}

	result := buf.Bytes()

	// Check if the result is aligned.
	if len(result)%encoding.TARGET_ALIGNMENT != 0 {
		log.Println("WARNING: prepareProofData() result is not aligned!")
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	// log.Printf("preppedAttestationData: %v",preppedAttestationData)

	attestationDataLen := len(preppedAttestationData)

	// Check if the attestation data length is too long.
	if attestationDataLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - attestationDataLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	methodLen := len(req.RequestMethod)
	// log.Printf("methodLen %v",methodLen)

	// Check if the method length is too long.
	if methodLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - methodLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	urlLen := len(req.Url)

	// Add the padding to the URL length.
	urlLenWithPadding := len(req.Url) + len(utils.GetPadding([]byte(req.Url), encoding.TARGET_ALIGNMENT))

	// Check if the URL length is too long.
	if urlLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - urlLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	selectorLen := len(req.Selector)

	// Add the padding to the selector length.
	selectorLenWithPadding := len(req.Selector) + len(utils.GetPadding([]byte(req.Selector), encoding.TARGET_ALIGNMENT))

	// Check if the selector length is too long.
	if selectorLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - selectorLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	headersLen := len(encodedHeaders)

	// Check if the headers length is too long.
	if headersLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - headersLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
	}

	optionalFieldsLen := len(encodedOptionalFields)

	// Check if the optional fields length is too long.
	if optionalFieldsLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - optionalFieldsLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.NewAppError(appErrors.ErrPreparationCriticalError)
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

	var startPos = 2

	var encodedPositions ProofPositionalInfo

	encodedPositions.Data = PositionInfo{
		Pos: startPos,
		Len: len(attestationDataBuffer) / encoding.TARGET_ALIGNMENT,
	}
	startPos += encodedPositions.Data.Len

	encodedPositions.Timestamp = PositionInfo{
		Pos: startPos,
		Len: 1,
	}
	startPos += encodedPositions.Timestamp.Len

	encodedPositions.StatusCode = PositionInfo{
		Pos: startPos,
		Len: 1,
	}
	startPos += encodedPositions.StatusCode.Len

	encodedPositions.Url = PositionInfo{
		Pos: startPos,
		Len: urlLenWithPadding / encoding.TARGET_ALIGNMENT,
	}
	startPos += encodedPositions.Url.Len

	encodedPositions.Selector = PositionInfo{
		Pos: startPos,
		Len: selectorLenWithPadding / encoding.TARGET_ALIGNMENT,
	}
	startPos += encodedPositions.Selector.Len

	encodedPositions.EncodingOptions = PositionInfo{
		Pos: startPos,
		Len: 1,
	}
	startPos += encodedPositions.EncodingOptions.Len

	encodedPositions.Method = PositionInfo{
		Pos: startPos,
		Len: 1,
	}
	startPos += encodedPositions.Method.Len

	encodedPositions.ResponseFormat = PositionInfo{
		Pos: startPos,
		Len: 1,
	}
	startPos += encodedPositions.ResponseFormat.Len

	encodedPositions.RequestHeaders = PositionInfo{
		Pos: startPos,
		Len: headersLen / encoding.TARGET_ALIGNMENT,
	}
	startPos += encodedPositions.RequestHeaders.Len

	encodedPositions.OptionalFields = PositionInfo{
		Pos: startPos,
		Len: optionalFieldsLen / encoding.TARGET_ALIGNMENT,
	}

	return result, encodedPositions, nil
}

// PrepareEncodedRequestProof prepares the encoded request proof.
func PrepareEncodedRequestProof(userData []byte, encodedPositions ProofPositionalInfo) ([]byte, *appErrors.AppError) {

	// Get the attestation data length and timestamp length.
	attestationDataLen := encodedPositions.Data.Len
	timestampLen := encodedPositions.Timestamp.Len

	// Calculate the end offset.
	endOffset := 32 + (attestationDataLen+timestampLen)*16

	// Check if the user data is too short.
	if endOffset > len(userData) {
		return nil, appErrors.NewAppError(appErrors.ErrUserDataTooShort)
	}

	// Set the user data to 0.
	for i := 32; i < endOffset; i++ {
		userData[i] = 0
	}

	return userData, nil
}
