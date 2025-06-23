package services

import (
	"bytes"
	"errors"
	"log"
	"math"
	"math/big"
	"strings"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

	encoding "github.com/zkportal/aleo-oracle-encoding"
	"github.com/zkportal/aleo-oracle-encoding/positionRecorder"
)

// PriceFeedBtcUrl, PriceFeedEthUrl, and PriceFeedAleoUrl are the URLs for the price feeds.
const (
	PriceFeedBtcUrl  = "price_feed: btc"
	PriceFeedEthUrl  = "price_feed: eth"
	PriceFeedAleoUrl = "price_feed: aleo"

	// AttestationDataSizeLimit is the size limit for the attestation data.
	AttestationDataSizeLimit = 1024 * 3
)

// getPadding gets the padding for the array.
func getPadding(arr []byte, alignment int) []byte {
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
func padStringToLength(str string, paddingChar byte, targetLength int) string {

	// Pad the string to the target length.
	return str + strings.Repeat(string(paddingChar), targetLength-len(str))
}

// prepareAttestationData prepares the attestation data.
func prepareAttestationData(attestationData string, encodingOptions *encoding.EncodingOptions) string {

	// Check the encoding option.
	switch encodingOptions.Value {
	case encoding.ENCODING_OPTION_STRING:
		// Pad the string to the target length.
		return padStringToLength(attestationData, 0x00, AttestationDataSizeLimit)
	case encoding.ENCODING_OPTION_FLOAT:
		// Check if the attestation data contains a dot.
		if strings.Contains(attestationData, ".") {
			return padStringToLength(attestationData, '0', math.MaxUint8)
		} else {
			// Pad the string to the target length.
			return padStringToLength(attestationData+".", '0', math.MaxUint8)
		}
	case encoding.ENCODING_OPTION_INT:
		// For integers we prepend zeroes instead of appending, that allows strconv to parse it no matter how many zeroes there are
		return padStringToLength("", '0', math.MaxUint8-len(attestationData)) + attestationData
	}

	return attestationData
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

// PrepareProofData prepares the proof data.
func PrepareProofData(statusCode int, attestationData string, timestamp int64, req AttestationRequest) ([]byte, ProofPositionalInfo, error) {

	// Prepare the attestation data.
	preppedAttestationData := attestationData

	// Check if the URL is a price feed.
	if req.Url != PriceFeedBtcUrl && req.Url != PriceFeedEthUrl && req.Url != PriceFeedAleoUrl {
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
		return nil, ProofPositionalInfo{}, appErrors.ErrEncodingAttestationData
	}

	// Write the attestation data to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, attestationDataBuffer); err != nil {
		log.Println("prepareProofData: failed to write attestation data to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingAttestationData
	}

	// Write the timestamp to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encoding.NumberToBytes(uint64(timestamp))); err != nil {
		log.Println("prepareProofData: failed to write timestamp to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingTimestamp
	}

	// Write the status code to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encoding.NumberToBytes(uint64(statusCode))); err != nil {
		log.Println("prepareProofData: failed to write status code to buffer, err = ", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingStatusCode
	}

	// Write the URL to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, []byte(req.Url)); err != nil {
		log.Println("prepareProofData: failed to write URL to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingUrl
	}

	// Write the selector to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, []byte(req.Selector)); err != nil {
		log.Println("prepareProofData: failed to write selector to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingSelector
	}

	// Encode the response format.
	responseFormat, err := encoding.EncodeResponseFormat(req.ResponseFormat)
	if err != nil {
		log.Println("prepareProofData: failed to encode response format, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrEncodingResponseFormat
	}

	// Write the response format to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, responseFormat); err != nil {
		log.Println("prepareProofData: failed to write response format to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingResponseFormat
	}

	// Write the request method to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, []byte(req.RequestMethod)); err != nil {
		log.Println("prepareProofData: failed to write request method to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingRequestMethod
	}

	// Encode the encoding options.
	encodingOptions, err := encoding.EncodeEncodingOptions(&req.EncodingOptions)
	if err != nil {
		log.Println("prepareProofData: failed to encode encoding options, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrEncodingEncodingOptions
	}

	// Write the encoding options to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encodingOptions); err != nil {
		log.Println("prepareProofData: failed to write encoding options to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingEncodingOptions
	}

	// Encode the request headers.
	encodedHeaders := encoding.EncodeHeaders(req.RequestHeaders)

	// Write the request headers to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encodedHeaders); err != nil {
		log.Println("prepareProofData: failed to write request headers to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingRequestHeaders
	}

	// write optional fields:
	// - html result type (exists only if response format is html)
	// - request content-type (can exist only if method is POST)
	// - request body (can exist only if method is POST)

	// Encode the optional fields.
	encodedOptionalFields, err := encoding.EncodeOptionalFields(req.HTMLResultType, req.RequestContentType, req.RequestBody)
	if err != nil {
		log.Println("prepareProofData: failed to encode optional fields, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrEncodingOptionalFields
	}

	// Write the optional fields to the buffer.
	if _, err = encoding.WriteWithPadding(recorder, encodedOptionalFields); err != nil {
		log.Println("prepareProofData: failed to write optional fields buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingOptionalFields
	}

	result := buf.Bytes()

	// Check if the result is aligned.
	if len(result)%encoding.TARGET_ALIGNMENT != 0 {
		log.Println("WARNING: prepareProofData() result is not aligned!")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	// log.Printf("preppedAttestationData: %v",preppedAttestationData)

	attestationDataLen := len(preppedAttestationData)

	// Check if the attestation data length is too long.
	if attestationDataLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - attestationDataLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	methodLen := len(req.RequestMethod)
	// log.Printf("methodLen %v",methodLen)

	// Check if the method length is too long.
	if methodLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - methodLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	urlLen := len(req.Url)

	// Add the padding to the URL length.
	urlLenWithPadding := len(req.Url) + len(getPadding([]byte(req.Url), encoding.TARGET_ALIGNMENT))

	// Check if the URL length is too long.
	if urlLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - urlLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	selectorLen := len(req.Selector)

	// Add the padding to the selector length.
	selectorLenWithPadding := len(req.Selector) + len(getPadding([]byte(req.Selector), encoding.TARGET_ALIGNMENT))

	// Check if the selector length is too long.
	if selectorLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - selectorLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	headersLen := len(encodedHeaders)

	// Check if the headers length is too long.
	if headersLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - headersLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	optionalFieldsLen := len(encodedOptionalFields)

	// Check if the optional fields length is too long.
	if optionalFieldsLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - optionalFieldsLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
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
func PrepareEncodedRequestProof(userData []byte, encodedPositions ProofPositionalInfo) ([]byte, error) {

	// Get the attestation data length and timestamp length.
	attestationDataLen := encodedPositions.Data.Len
	timestampLen := encodedPositions.Timestamp.Len

	// Calculate the end offset.
	endOffset := 32 + (attestationDataLen+timestampLen)*16

	// Check if the user data is too short.
	if endOffset > len(userData) {
		return nil, appErrors.ErrUserDataTooShort
	}

	// Set the user data to 0.
	for i := 32; i < endOffset; i++ {
		userData[i] = 0
	}

	return userData, nil
}
