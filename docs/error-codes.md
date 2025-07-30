# Error Codes Documentation

This document provides a comprehensive reference for all error codes used in the Aleo Oracle Notarization Backend. Each error code is categorized by functionality and includes detailed descriptions, usage examples, and troubleshooting guidance.

## Error Code Structure

All error codes follow a structured numbering system:

- **1000-1999**: Validation Errors
- **2000-2999**: Enclave/SGX Errors  
- **3000-3999**: Attestation Errors
- **4000-4999**: Data Extraction Errors
- **5000-5999**: Encoding Errors
- **6000-6999**: Price Feed Errors
- **7000-7999**: Request/Response Errors
- **8000-8999**: Internal Errors

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
| `1008` | `ErrInvalidRequestBody` | Request body is not allowed with GET requestMethod | 400 |
| `1019` | `ErrMissingRequestContentType` | Request content type is required with POST requestMethod | 400 |
| `1020` | `ErrInvalidRequestContentType` | Request content type is not allowed with GET requestMethod | 400 |

### Response Format Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1009` | `ErrInvalidResponseFormat` | Response format must be html or json | 400 |
| `1010` | `ErrMissingHTMLResultType` | HTML result type required for html format | 400 |
| `1011` | `ErrInvalidHTMLResultType` | HTML result type must be element or value | 400 |
| `1013` | `ErrInvalidHTMLResultTypeForJSONResponse` | HTML result type is not allowed with json responseFormat | 400 |

### Encoding Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1012` | `ErrInvalidEncodingOptionForHTMLResultType` | Expected encodingOptions.value to be string with htmlResultType element | 400 |
| `1014` | `ErrMissingEncodingValue` | Encoding options value is required | 400 |
| `1015` | `ErrInvalidEncodingOption` | Invalid encoding option. expected: string/float/int | 400 |
| `1017` | `ErrMissingEncodingPrecision` | Encoding options precision is required for float encoding | 400 |
| `1018` | `ErrInvalidEncodingPrecision` | Encoding options precision should be 0 for int/string encoding and greater than 0 and less than 12 for float encoding | 400 |

### Security Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1016` | `ErrTargetNotWhitelisted` | Attestation target is not whitelisted | 403 |

### URL Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1022` | `ErrInvalidTargetURL` | URL should not include a scheme. Please remove https:// or http:// from your url | 400 |
| `1028` | `ErrInvalidURL` | Invalid URL format | 400 |

### Max Parameter Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1023` | `ErrMissingMaxParameter` | Missing max search parameter | 400 |
| `1024` | `ErrInvalidMaxParameter` | Invalid max search parameter | 400 |
| `1025` | `ErrInvalidMaxValue` | Expected max search parameter to be a number 2-2^127 | 400 |
| `1026` | `ErrInvalidMaxValueFormat` | Invalid max value format | 400 |

### Attestation Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1027` | `ErrInvalidAttestationData` | Attestation data is invalid | 400 |

### Price Feed Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `1021` | `ErrInvalidSelector` | Selector expected to be weightedAvgPrice for price feed requests | 400 |
| `1029` | `ErrInvalidResponseFormatForPriceFeed` | Response format expected to be json for price feed requests | 400 |
| `1030` | `ErrInvalidEncodingOptionForPriceFeed` | Invalid encoding option. expected: float for price feed requests | 400 |
| `1031` | `ErrInvalidRequestMethodForPriceFeed` | Request method expected to be GET for price feed requests | 400 |
| `1032` | `ErrInvalidSelectorForPriceFeed` | Selector expected to be weightedAvgPrice for price feed requests | 400 |

## 2. ENCLAVE ERRORS (2000-2999)

Enclave errors occur during SGX enclave operations, including quote generation, target info handling, and report data management.

### Target Info Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `2001` | `ErrReadingTargetInfo` | Failed to read the target info | 500 |
| `2005` | `ErrWrittingTargetInfo` | Failed to write the target info | 500 |

### Report Data Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `2002` | `ErrWrittingReportData` | Failed to write the report data | 500 |
| `2007` | `ErrReadingReport` | Failed to read the report | 500 |

### Quote Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `2003` | `ErrGeneratingQuote` | Failed to generate the quote | 500 |
| `2004` | `ErrReadingQuote` | Failed to read the quote | 500 |
| `2006` | `ErrWrappingQuote` | Failed to wrap quote in openenclave format | 500 |

### SGX Report Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `2008` | `ErrInvalidSGXReportSize` | Invalid SGX report size | 500 |
| `2009` | `ErrParsingSGXReport` | Failed to parse SGX report | 500 |
| `2010` | `ErrEmptyQuote` | Empty quote | 500 |

## 3. ATTESTATION ERRORS (3000-3999)

Attestation errors occur during the process of creating and formatting attestation data, proofs, and signatures.

### Data Preparation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `3001` | `ErrPreparingHashMessage` | Failed to prepare hash message for oracle data | 500 |
| `3002` | `ErrPreparingProofData` | Failed to prepare proof data | 500 |

### Data Formatting

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `3003` | `ErrFormattingProofData` | Failed to format proof data | 500 |
| `3005` | `ErrFormattingEncodedProofData` | Failed to format encoded proof data | 500 |
| `3008` | `ErrFormattingQuote` | Failed to format quote | 500 |

### Hashing Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `3004` | `ErrCreatingAttestationHash` | Failed to create attestation hash | 500 |
| `3006` | `ErrCreatingRequestHash` | Failed to create request hash | 500 |
| `3007` | `ErrCreatingTimestampedRequestHash` | Failed to create timestamped request hash | 500 |
| `3009` | `ErrHashingReport` | Failed to hash the oracle report | 500 |

### Signature Generation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `3010` | `ErrGeneratingSignature` | Failed to generate signature | 500 |

### Aleo Context

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `7001` | `ErrAleoContext` | Failed to initialize Aleo context | 500 |

## 4. DATA EXTRACTION ERRORS (4000-4999)

Data extraction errors occur during HTTP requests, HTML/JSON parsing, and selector operations.

### HTTP Request Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4001` | `ErrInvalidHTTPRequest` | Invalid HTTP request | 400 |
| `4002` | `ErrFetchingData` | Failed to fetch the data from the provided endpoint | 500 |
| `4003` | `ErrInvalidStatusCode` | Invalid status code returned from endpoint | 500 |

### HTML Processing Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4004` | `ErrReadingHTMLContent` | Failed to read HTML content from target url | 500 |
| `4005` | `ErrParsingHTMLContent` | Failed to parse HTML content from target url | 500 |

### JSON Processing Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4006` | `ErrReadingJSONResponse` | Failed to read the json response from target url | 500 |
| `4007` | `ErrDecodingJSONResponse` | Failed to decode JSON response from target url | 500 |

### Selector Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4008` | `ErrSelectorNotFound` | Selector not found | 404 |

### Price Feed Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4009` | `ErrUnsupportedPriceFeedURL` | Unsupported price feed URL | 400 |

### Data Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `4010` | `ErrAttestationDataTooLarge` | Attestation data too large | 413 |
| `4011` | `ErrParsingFloatValue` | Extracted value expected to be float but failed to parse as float | 500 |
| `4012` | `ErrParsingIntValue` | Extracted value expected to be int but failed to parse as int | 500 |
| `4013` | `ErrEmptyAttestationData` | Extracted attestation data is empty | 400 |

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

### Meta Header Preparation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5006` | `ErrPreparingMetaHeader` | Error while preparing meta header | 500 |

### Specific Field Encoding

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5008` | `ErrWrittingTimestamp` | Failed to write timestamp to buffer | 500 |
| `5009` | `ErrWrittingStatusCode` | Failed to write status code to buffer | 500 |
| `5010` | `ErrWrittingUrl` | Failed to write url to buffer | 500 |
| `5011` | `ErrWrittingSelector` | Failed to write selector to buffer | 500 |
| `5013` | `ErrWrittingRequestMethod` | Failed to write request method to buffer | 500 |

### User Data Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `5017` | `ErrUserDataTooShort` | User data too short for expected zeroing | 400 |

## 6. PRICE FEED ERRORS (6000-6999)

Price feed errors occur during cryptocurrency price feed operations and exchange API interactions.

### Token Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6001` | `ErrTokenNotSupported` | Token not supported. Supported tokens: BTC, ETH, ALEO | 400 |

### Exchange Configuration

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6002` | `ErrExchangeNotConfigured` | Exchange not configured | 500 |
| `6003` | `ErrSymbolNotConfigured` | Symbol not configured | 500 |
| `6004` | `ErrExchangeNotSupported` | Exchange not supported | 400 |

### API Operations

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6005` | `ErrCreatingExchangeRequest` | Failed to create exchange request | 500 |
| `6006` | `ErrFetchingFromExchange` | Failed to fetch from exchange | 500 |
| `6007` | `ErrExchangeInvalidStatusCode` | Invalid status code returned from exchange | 500 |

### Response Processing

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6008` | `ErrReadingExchangeResponse` | Failed to read exchange response | 500 |
| `6009` | `ErrDecodingExchangeResponse` | Failed to decode exchange response | 500 |
| `6010` | `ErrParsingExchangeResponse` | Failed to parse exchange response | 500 |

### Data Validation

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6011` | `ErrMissingDataInResponse` | Missing data in response | 404 |
| `6012` | `ErrParsingPrice` | Failed to parse price | 500 |
| `6013` | `ErrParsingVolume` | Failed to parse volume | 500 |

### Data Availability

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6014` | `ErrInsufficientExchangeData` | Insufficient data from exchanges | 503 |

### Data Encoding

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6015` | `ErrEncodingPriceFeedData` | Failed to encode price feed data | 500 |

### Configuration

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `6016` | `ErrNoTradingPairsConfigured` | No trading pairs configured for token | 500 |

## 7. REQUEST/RESPONSE ERRORS (7000-7999)

Request and response errors occur during HTTP request processing and response handling.

### Request Processing

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `7001` | `ErrRequestBodyTooLarge` | Payload exceeds the allowed size limit | 413 |
| `7002` | `ErrReadingRequestBody` | Failed to read the request body | 400 |
| `7003` | `ErrInvalidContentType` | Invalid content type, expected application/json | 400 |
| `7004` | `ErrDecodingRequestBody` | Failed to decode request body, invalid request structure | 400 |

## 8. INTERNAL ERRORS (8000-8999)

Internal errors occur due to unexpected system failures or configuration issues.

### System Errors

| Code | Error Name | Description | HTTP Status |
|------|------------|-------------|-------------|
| `8001` | `ErrInternal` | Unexpected failure occurred | 500 |
| `8002` | `ErrGeneratingRandomNumber` | Failed to generate random number | 500 |
| `8003` | `ErrJSONEncoding` | Failed to encode data to JSON | 500 |

## Usage Examples

### Creating Errors

```go
import appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

// Basic error creation
err := appErrors.ErrMissingURL

// Error with custom response status
err := appErrors.ErrInvalidRequestMethod.WithResponseStatusCode(400)

// Error with additional details
err := appErrors.ErrFetchingData.WithDetails("Connection timeout after 30 seconds")
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
        case 1016: // ErrTargetNotWhitelisted
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

#### Price Feed Errors (6000-6999)
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