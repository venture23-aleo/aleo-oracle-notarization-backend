# Error Codes Documentation

This document provides a comprehensive reference for all error codes used in the Aleo Oracle Notarization Backend. Each error code is categorized by functionality and includes detailed descriptions, usage examples, and troubleshooting guidance.

## Error Code Structure

All error codes follow a structured numbering system:

- **1000-1999**: Validation Errors
- **2000-2999**: Enclave/SGX Errors  
- **3000-3999**: Attestation Errors
- **4000-4999**: Data Extraction Errors
- **5000-5999**: Encoding Errors
- **6000-6999**: Exchange/Price Feed Errors
- **7000-7999**: Aleo Context Errors

## Error Response Format

All errors are returned in a consistent JSON format:

```json
{
  "errorCode": 1001,
  "errorMessage": "validation error: url is required",
  "errorDetails": "URL field was empty",
  "responseStatusCode": 400,
  "requestId": "req-123"
}
```

## 1. VALIDATION ERRORS (1000-1999)

Validation errors occur when input data fails to meet the required format, structure, or business rules.

### Core Validation Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1001` | `ErrMissingURL` | URL parameter is required but missing | 400 |
| `1002` | `ErrMissingRequestMethod` | Request method (GET/POST) is required | 400 |
| `1003` | `ErrMissingResponseFormat` | Response format (html/json) is required | 400 |
| `1004` | `ErrMissingSelector` | CSS selector for data extraction is required | 400 |
| `1005` | `ErrMissingEncodingOption` | Encoding option value is required | 400 |

### Request Method Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1006` | `ErrInvalidRequestMethod` | Request method must be GET or POST | 400 |
| `1007` | `ErrMissingRequestBody` | Request body is required for POST requests | 400 |

### Response Format Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1008` | `ErrInvalidResponseFormat` | Response format must be html or json | 400 |
| `1009` | `ErrMissingHTMLResultType` | HTML result type required for html format | 400 |
| `1010` | `ErrInvalidHTMLResultType` | HTML result type must be element or value | 400 |

### Encoding Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1011` | `ErrInvalidEncodingOption` | Invalid encoding option provided | 400 |
| `1014` | `ErrInvalidEncodingPrecision` | Encoding precision is out of valid range | 400 |

### Security Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1012` | `ErrUnacceptedDomain` | Target domain is not in whitelist | 403 |

### Request Structure Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1013` | `ErrInvalidRequestData` | Request structure is invalid or malformed | 400 |

### Max Parameter Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1015` | `ErrMissingMaxParameter` | Max parameter is required | 400 |
| `1016` | `ErrInvalidMaxParameter` | Max parameter value is invalid | 400 |
| `1017` | `ErrInvalidMaxValue` | Max value is invalid | 400 |
| `1018` | `ErrInvalidMaxValueFormat` | Max value format is invalid | 400 |
| `1019` | `ErrInvalidMaxValueRange` | Max value is out of acceptable range | 400 |

### System Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1020` | `ErrGeneratingRandomNumber` | Failed to generate random number | 500 |

## 2. ENCLAVE ERRORS (2000-2999)

Enclave errors occur during SGX enclave operations, including quote generation, target info handling, and report data management.

### Target Info Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `2001` | `ErrReadingTargetInfo` | Failed to read target info from enclave | 500 |
| `2005` | `ErrWrittingTargetInfo` | Failed to write target info to enclave | 500 |

### Report Data Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `2002` | `ErrWrittingReportData` | Failed to write report data to enclave | 500 |
| `2007` | `ErrReadingReport` | Failed to read report from enclave | 500 |

### Quote Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `2003` | `ErrGeneratingQuote` | Failed to generate SGX quote | 500 |
| `2004` | `ErrReadingQuote` | Failed to read quote from enclave | 500 |
| `2006` | `ErrWrappingQuote` | Failed to wrap quote in OpenEnclave format | 500 |

## 3. ATTESTATION ERRORS (3000-3999)

Attestation errors occur during the process of creating and formatting attestation data, proofs, and signatures.

### Oracle Data Preparation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `3001` | `ErrPreparingOracleData` | Failed to prepare oracle data | 500 |
| `3003` | `ErrPreparingProofData` | Failed to prepare proof data | 500 |
| `3006` | `ErrPreparingEncodedProof` | Failed to prepare encoded request proof | 500 |

### Data Formatting

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `3004` | `ErrFormattingProofData` | Failed to format proof data | 500 |
| `3007` | `ErrFormattingEncodedProofData` | Failed to format encoded proof data | 500 |
| `3011` | `ErrFormattingQuote` | Failed to format quote | 500 |
| `3014` | `ErrDecodingQuote` | Failed to decode quote | 500 |

### Hashing Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `3002` | `ErrMessageHashing` | Failed to hash message for oracle data | 500 |
| `3009` | `ErrCreatingRequestHash` | Failed to create request hash | 500 |
| `3010` | `ErrCreatingTimestampedRequestHash` | Failed to create timestamped request hash | 500 |
| `3012` | `ErrReportHashing` | Failed to hash oracle report | 500 |

### Attestation Generation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `3005` | `ErrGeneratingAttestationHash` | Failed to generate attestation hash | 500 |
| `3013` | `ErrGeneratingSignature` | Failed to generate signature | 500 |

### Data Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `3008` | `ErrUserDataTooShort` | User data too short for expected zeroing | 400 |

## 4. DATA EXTRACTION ERRORS (4000-4999)

Data extraction errors occur during HTTP requests, HTML/JSON parsing, and selector operations.

### HTTP Request Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4001` | `ErrInvalidHTTPRequest` | Invalid HTTP request format | 400 |
| `4002` | `ErrFetchingData` | Failed to fetch data from endpoint | 500 |

### HTML Processing Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4003` | `ErrReadingHTMLContent` | Failed to read HTML content | 500 |
| `4004` | `ErrParsingHTMLContent` | Failed to parse HTML content | 500 |

### Selector Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4005` | `ErrSelectorNotFound` | CSS selector not found in content | 404 |
| `4010` | `ErrInvalidSelectorPart` | Invalid selector part | 400 |

### JSON Processing Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4008` | `ErrJSONDecoding` | Failed to decode JSON response | 500 |
| `4009` | `ErrJSONEncoding` | Failed to encode data to JSON | 500 |

### Data Structure Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4006` | `ErrInvalidMap` | Expected map but got different type | 500 |
| `4007` | `ErrKeyNotFound` | Required key not found in data | 404 |
| `4011` | `ErrExpectedArray` | Expected array at specified key | 500 |
| `4012` | `ErrIndexOutOfBound` | Array index out of bounds | 500 |

### Specialized Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4013` | `ErrUnsupportedPriceFeedURL` | Unsupported price feed URL | 400 |
| `4014` | `ErrAttestationDataTooLarge` | Attestation data exceeds size limit | 413 |

## 5. ENCODING ERRORS (5000-5999)

Encoding errors occur during data encoding, buffer writing, and format conversion operations.

### Attestation Data Encoding

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5001` | `ErrEncodingAttestationData` | Failed to encode attestation data | 500 |
| `5007` | `ErrWrittingAttestationData` | Failed to write attestation data to buffer | 500 |

### Response Format Encoding

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5002` | `ErrEncodingResponseFormat` | Failed to encode response format | 500 |
| `5012` | `ErrWrittingResponseFormat` | Failed to write response format to buffer | 500 |

### Encoding Options

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5003` | `ErrEncodingEncodingOptions` | Failed to encode encoding options | 500 |
| `5014` | `ErrWrittingEncodingOptions` | Failed to write encoding options to buffer | 500 |

### Header Encoding

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5004` | `ErrEncodingHeaders` | Failed to encode headers | 500 |
| `5015` | `ErrWrittingRequestHeaders` | Failed to write request headers to buffer | 500 |

### Optional Fields Encoding

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5005` | `ErrEncodingOptionalFields` | Failed to encode optional fields | 500 |
| `5016` | `ErrWrittingOptionalFields` | Failed to write optional headers to buffer | 500 |

### Specific Field Encoding

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5008` | `ErrWrittingTimestamp` | Failed to write timestamp to buffer | 500 |
| `5009` | `ErrWrittingStatusCode` | Failed to write status code to buffer | 500 |
| `5010` | `ErrWrittingUrl` | Failed to write URL to buffer | 500 |
| `5011` | `ErrWrittingSelector` | Failed to write selector to buffer | 500 |
| `5013` | `ErrWrittingRequestMethod` | Failed to write request method to buffer | 500 |

### Critical Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5006` | `ErrPreparationCriticalError` | Critical error while preparing data | 500 |

## 6. EXCHANGE ERRORS (6000-6999)

Exchange errors occur during cryptocurrency price feed operations and exchange API interactions.

### Token Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6001` | `ErrMissingToken` | Token parameter is required | 400 |
| `6002` | `ErrInvalidToken` | Invalid token (supported: BTC, ETH, ALEO) | 400 |

### Exchange Configuration

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6004` | `ErrExchangeNotConfigured` | Exchange not configured | 500 |
| `6006` | `ErrUnsupportedExchange` | Unsupported exchange | 400 |

### API Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6003` | `ErrPriceFeedFailed` | Failed to get price feed data | 500 |
| `6007` | `ErrExchangeFetchFailed` | Failed to fetch from exchange | 500 |
| `6008` | `ErrExchangeInvalidStatusCode` | Exchange returned invalid status code | 500 |

### Data Processing

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6009` | `ErrExchangeResponseDecodeFailed` | Failed to decode exchange response | 500 |
| `6010` | `ErrExchangeResponseParseFailed` | Failed to parse exchange response | 500 |
| `6016` | `ErrNoDataInResponse` | No data in exchange response | 404 |

### Symbol Support

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6005` | `ErrSymbolNotSupportedByExchange` | Symbol not supported by exchange | 400 |

### Data Format Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6011` | `ErrInvalidPriceFormat` | Invalid price format | 500 |
| `6012` | `ErrInvalidVolumeFormat` | Invalid volume format | 500 |
| `6013` | `ErrInvalidExchangeResponseFormat` | Invalid exchange response format | 500 |
| `6014` | `ErrInvalidDataFormat` | Invalid data format | 500 |
| `6015` | `ErrInvalidItemFormat` | Invalid item format | 500 |

### Data Parsing

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6017` | `ErrPriceParseFailed` | Failed to parse price | 500 |
| `6018` | `ErrVolumeParseFailed` | Failed to parse volume | 500 |

### Data Availability

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6019` | `ErrInsufficientExchangeData` | Insufficient data from exchanges | 503 |

## 7. ALEO CONTEXT ERRORS (7000-7999)

Aleo context errors occur during Aleo blockchain context initialization and operations.

### Context Initialization

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `7001` | `ErrAleoContext` | Failed to initialize Aleo context | 500 |

## Usage Examples

### Creating Errors

```go
import appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

// Basic error creation
err := appErrors.NewAppError(appErrors.ErrMissingURL)

// Error with custom response status
err := appErrors.NewAppErrorWithResponseStatus(appErrors.ErrInvalidRequestMethod, 400)

// Error with additional details
err := appErrors.NewAppErrorWithDetails(appErrors.ErrFetchingData, "Connection timeout after 30 seconds")
```

### Error Handling

```go
func handleAttestationRequest(req *AttestationRequest) error {
    // Validate request
    if err := req.Validate(); err != nil {
        switch err.Code {
        case 1001: // ErrMissingURL
            log.Error("URL parameter missing", "error", err)
            return err
        case 1006: // ErrInvalidRequestMethod
            log.Error("Invalid request method", "method", req.RequestMethod)
            return err
        case 1012: // ErrUnacceptedDomain
            log.Error("Domain not whitelisted", "domain", req.URL)
            return err
        default:
            log.Error("Validation error", "code", err.Code, "message", err.Message)
            return err
        }
    }
    
    return nil
}
```

### HTTP Response Handling

```go
func writeErrorResponse(w http.ResponseWriter, err *appErrors.AppError) {
    statusCode := err.ResponseStatusCode
    if statusCode == 0 {
        statusCode = 500 // Default to internal server error
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    response := map[string]interface{}{
        "error": true,
        "errorCode": err.Code,
        "errorMessage": err.Message,
    }
    
    if err.Details != "" {
        response["errorDetails"] = err.Details
    }
    
    json.NewEncoder(w).Encode(response)
}
```

## Troubleshooting Guide

### Common Error Scenarios

#### Validation Errors (1000-1999)
- **Cause**: Invalid or missing input parameters
- **Solution**: Check request format and required fields
- **Prevention**: Implement client-side validation

#### Enclave Errors (2000-2999)
- **Cause**: SGX hardware issues or enclave configuration problems
- **Solution**: Verify SGX drivers and enclave setup
- **Prevention**: Regular SGX health checks

#### Data Extraction Errors (4000-4999)
- **Cause**: Network issues or changes in target website structure
- **Solution**: Check network connectivity and update selectors
- **Prevention**: Implement retry mechanisms and fallback data sources

#### Exchange Errors (6000-6999)
- **Cause**: Exchange API changes or rate limiting
- **Solution**: Update exchange configurations and implement rate limiting
- **Prevention**: Monitor exchange API status and implement circuit breakers

### Error Monitoring

```go
// Log errors with structured logging
logger.Error("Application error occurred",
    "errorCode", err.Code,
    "errorMessage", err.Message,
    "errorDetails", err.Details,
    "requestId", requestID,
    "timestamp", time.Now(),
)
```

### Error Metrics

Track error rates by category:

```go
// Prometheus metrics
validationErrors.WithLabelValues("missing_url").Inc()
enclaveErrors.WithLabelValues("quote_generation").Inc()
dataExtractionErrors.WithLabelValues("http_request").Inc()
```

## Best Practices

1. **Always include error codes** in responses for client handling
2. **Log errors with context** including request ID and timestamp
3. **Implement proper error handling** at each layer
4. **Use appropriate HTTP status codes** for different error types
5. **Monitor error rates** to identify system issues
6. **Provide meaningful error messages** for debugging
7. **Implement retry mechanisms** for transient errors
8. **Use circuit breakers** for external service calls

## Error Code Maintenance

When adding new error codes:

1. **Follow the numbering scheme** for the appropriate category
2. **Add comprehensive documentation** including description and HTTP status
3. **Include usage examples** in tests
4. **Update this documentation** with new error codes
5. **Consider backward compatibility** when modifying existing codes 