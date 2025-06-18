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

const (
	PriceFeedBtcUrl  = "price_feed: btc"
	PriceFeedEthUrl  = "price_feed: eth"
	PriceFeedAleoUrl = "price_feed: aleo"

	AttestationDataSizeLimit = 1024 * 3
)

func getPadding(arr []byte, alignment int) []byte {
	var paddingSize int
	overflow := len(arr) % alignment
	if overflow != 0 {
		paddingSize = alignment - overflow
	} else {
		paddingSize = 0
	}
	padding := make([]byte, paddingSize)
	return padding
}

func padStringToLength(str string, paddingChar byte, targetLength int) string {
	return str + strings.Repeat(string(paddingChar), targetLength-len(str))
}

func prepareAttestationData(attestationData string, encodingOptions *encoding.EncodingOptions) string {
	switch encodingOptions.Value {
	case encoding.ENCODING_OPTION_STRING:
		return padStringToLength(attestationData, 0x00, AttestationDataSizeLimit)
	case encoding.ENCODING_OPTION_FLOAT:
		if strings.Contains(attestationData, ".") {
			return padStringToLength(attestationData, '0', math.MaxUint8)
		} else {
			return padStringToLength(attestationData+".", '0', math.MaxUint8)
		}
	case encoding.ENCODING_OPTION_INT:
		// for integers we prepend zeroes instead of appending, that allows strconv to parse it no matter how many zeroes there are
		return padStringToLength("", '0', math.MaxUint8-len(attestationData)) + attestationData
	}

	return attestationData
}

func SliceToU128(buf []byte) (*big.Int, error) {
	if len(buf) != 16 {
		return nil, errors.New("cannot convert slice to u128: invalid size")
	}

	result := big.NewInt(0)

	for idx, b := range buf {
		bigByte := big.NewInt(int64(b))
		bigByte.Lsh(bigByte, 8*uint(idx))
		result.Add(result, bigByte)
	}

	return result, nil
}

func PrepareProofData(statusCode int, attestationData string, timestamp int64, req AttestationRequest) ([]byte, ProofPositionalInfo, error) {
	preppedAttestationData := attestationData

	if req.Url != PriceFeedBtcUrl && req.Url != PriceFeedEthUrl && req.Url != PriceFeedAleoUrl {
		preppedAttestationData = prepareAttestationData(attestationData, &req.EncodingOptions)
	}

	var buf bytes.Buffer

	// information about the positions and lengths of all the encoded elements
	recorder := positionRecorder.NewPositionRecorder(&buf, encoding.TARGET_ALIGNMENT)

	// write an empty meta header
	encoding.WriteWithPadding(recorder, make([]byte, encoding.TARGET_ALIGNMENT*2))

	// write attestationData
	attestationDataBuffer, err := encoding.EncodeAttestationData(preppedAttestationData, &req.EncodingOptions)
	if err != nil {
		log.Println("prepareProofData: failed to encode attestation data, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrEncodingAttestationData
	}

	// attestationDataInt128, err := SliceToU128(attestationDataBuffer)

	// log.Printf("Attestation data int 128: %v",attestationDataInt128)

	if _, err = encoding.WriteWithPadding(recorder, attestationDataBuffer); err != nil {
		log.Println("prepareProofData: failed to write attestation data to buffer, err =", err)
		return nil,ProofPositionalInfo{}, appErrors.ErrWrittingAttestationData
	}

	// write timestamp
	if _, err = encoding.WriteWithPadding(recorder, encoding.NumberToBytes(uint64(timestamp))); err != nil {
		log.Println("prepareProofData: failed to write timestamp to buffer, err =", err)
		return nil,ProofPositionalInfo{}, appErrors.ErrWrittingTimestamp
	}

	// write status code
	if _, err = encoding.WriteWithPadding(recorder, encoding.NumberToBytes(uint64(statusCode))); err != nil {
		log.Println("prepareProofData: failed to write status code to buffer, err = ", err)
		return nil, ProofPositionalInfo{},appErrors.ErrWrittingStatusCode
	}

	// write url
	if _, err = encoding.WriteWithPadding(recorder, []byte(req.Url)); err != nil {
		log.Println("prepareProofData: failed to write URL to buffer, err =", err)
		return nil,ProofPositionalInfo{}, appErrors.ErrWrittingUrl
	}

	// write selector
	if _, err = encoding.WriteWithPadding(recorder, []byte(req.Selector)); err != nil {
		log.Println("prepareProofData: failed to write selector to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingSelector
	}

	// write response format
	responseFormat, err := encoding.EncodeResponseFormat(req.ResponseFormat)
	if err != nil {
		log.Println("prepareProofData: failed to encode response format, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrEncodingResponseFormat
	}

	if _, err = encoding.WriteWithPadding(recorder, responseFormat); err != nil {
		log.Println("prepareProofData: failed to write response format to buffer, err =", err)
		return nil,ProofPositionalInfo{}, appErrors.ErrWrittingResponseFormat
	}

	// write request method
	if _, err = encoding.WriteWithPadding(recorder, []byte(req.RequestMethod)); err != nil {
		log.Println("prepareProofData: failed to write request method to buffer, err =", err)
		return nil,ProofPositionalInfo{}, appErrors.ErrWrittingRequestMethod
	}

	// write encoding options
	encodingOptions, err := encoding.EncodeEncodingOptions(&req.EncodingOptions)
	if err != nil {
		log.Println("prepareProofData: failed to encode encoding options, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrEncodingEncodingOptions
	}

	if _, err = encoding.WriteWithPadding(recorder, encodingOptions); err != nil {
		log.Println("prepareProofData: failed to write encoding options to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingEncodingOptions
	}

	// write request headers
	encodedHeaders := encoding.EncodeHeaders(req.RequestHeaders)
	if _, err = encoding.WriteWithPadding(recorder, encodedHeaders); err != nil {
		log.Println("prepareProofData: failed to write request headers to buffer, err =", err)
		return nil, ProofPositionalInfo{}, appErrors.ErrWrittingRequestHeaders
	}

	// write optional fields:
	// - html result type (exists only if response format is html)
	// - request content-type (can exist only if method is POST)
	// - request body (can exist only if method is POST)
	encodedOptionalFields, err := encoding.EncodeOptionalFields(req.HTMLResultType, req.RequestContentType, req.RequestBody)
	if err != nil {
		log.Println("prepareProofDat: failed to write request's optional fields, err =", err)
		return nil,ProofPositionalInfo{}, appErrors.ErrEncodingOptionalFields
	}
	if _, err = encoding.WriteWithPadding(recorder, encodedOptionalFields); err != nil {
		log.Println("prepareProofData: failed to write optional fields buffer, err =", err)
		return nil,ProofPositionalInfo{}, appErrors.ErrWrittingOptionalFields
	}

	result := buf.Bytes()
	// failsafe
	if len(result)%encoding.TARGET_ALIGNMENT != 0 {
		log.Println("WARNING: prepareProofData() result is not aligned!")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	// log.Printf("preppedAttestationData: %v",preppedAttestationData)

	attestationDataLen := len(preppedAttestationData)

	if attestationDataLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - attestationDataLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	methodLen := len(req.RequestMethod)
	// log.Printf("methodLen %v",methodLen)

	if methodLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - methodLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	urlLen := len(req.Url)
	urlLenWithPadding := len(req.Url) + len(getPadding([]byte(req.Url), encoding.TARGET_ALIGNMENT))

	if urlLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - urlLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	selectorLen := len(req.Selector)
	selectorLenWithPadding := len(req.Selector) + len(getPadding([]byte(req.Selector), encoding.TARGET_ALIGNMENT))
	if selectorLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - selectorLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	headersLen := len(encodedHeaders)
	if headersLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - headersLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	optionalFieldsLen := len(encodedOptionalFields)
	if optionalFieldsLen > math.MaxUint16 {
		log.Println("Warning: cannot create encoded data meta header - optionalFieldsLen is too long")
		return nil, ProofPositionalInfo{}, appErrors.ErrPreparationCriticalError
	}

	// log.Printf("attestationDataLen: %v",attestationDataLen)
	// log.Printf("methodLen: %v",methodLen)
	// log.Printf("urlLen: %v",urlLen)
	// log.Printf("selectorLen: %v",selectorLen)
	// log.Printf("headersLen: %v",headersLen)

	// fill the empty meta header with the actual content
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

// Setting byte positions of attestation data and timestamp to 0 
func PrepareEncodedRequestProof(userData []byte, encodedPositions ProofPositionalInfo) ([]byte, error) {

	// attestationDataPos := encodedPositions.Data.Pos

	// timestampPos := encodedPositions.Timestamp.Pos
	attestationDataLen := encodedPositions.Data.Len
	timestampLen := encodedPositions.Timestamp.Len

	// attestationDataSize := binary.LittleEndian.Uint16(userData[0:2])
	// timestampSize := binary.LittleEndian.Uint16(userData[2:4])

	endOffset := 32 + (attestationDataLen +  timestampLen) * 16

	if endOffset > len(userData) {
		return nil, appErrors.ErrUserDataTooShort
	}

	for i := 32; i < endOffset; i++ {
		userData[i] = 0
	}

	return userData, nil
}
