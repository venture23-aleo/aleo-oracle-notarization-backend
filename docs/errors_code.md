# Error Codes Documentation

This document outlines the error code ranges and categories used in the Aleo Oracle Notarization Backend.

## Error Code Ranges

### 1. VALIDATION ERRORS (1000-1999)
**Purpose:** All input validation errors

**Categories:**
- URL validation errors
- Request method validation
- Response format validation
- Selector validation
- Encoding precision validation

**Common Error Codes:**
- `1001` - Invalid URL format
- `1002` - Unsupported request method
- `1003` - Invalid response format
- `1004` - Selector validation failed
- `1005` - Encoding precision out of range

---

### 2. ENCLAVE ERRORS (2000-2999)
**Purpose:** All SGX/enclave-related errors

**Categories:**
- Quote generation errors
- Reading operations errors
- Writing operations errors
- Target info operations errors

**Common Error Codes:**
- `2001` - Quote generation failed
- `2002` - Target info reading failed
- `2003` - Report data writing failed
- `2004` - Enclave initialization failed
- `2005` - SGX device access error

---

### 3. ATTESTATION ERRORS (3000-3999)
**Purpose:** All attestation process errors

**Categories:**
- Oracle data preparation errors
- Proof data generation errors
- Hashing operation errors
- Signature generation errors
- Quote formatting errors

**Common Error Codes:**
- `3001` - Oracle data preparation failed
- `3002` - Proof data generation failed
- `3003` - Hashing operation failed
- `3004` - Signature generation failed
- `3005` - Quote formatting failed

---

### 4. DATA EXTRACTION ERRORS (4000-4999)
**Purpose:** All data fetching and parsing errors

**Categories:**
- HTTP request errors
- HTML parsing errors
- JSON parsing errors
- Selector operation errors
- Key lookup errors

**Common Error Codes:**
- `4001` - HTTP request failed
- `4002` - HTML parsing failed
- `4003` - JSON parsing failed
- `4004` - Selector operation failed
- `4005` - Key lookup failed
- `4006` - Network timeout
- `4007` - Rate limit exceeded

---

### 5. ENCODING ERRORS (5000-5999)
**Purpose:** All encoding-related errors

**Categories:**
- Buffer writing errors
- Data encoding failures
- Format conversion errors

**Common Error Codes:**
- `5001` - Buffer writing failed
- `5002` - Data encoding failed
- `5003` - Base64 encoding failed
- `5004` - JSON encoding failed
- `5005` - Format conversion failed

---

### 6. EXCHANGE ERRORS (6000-6999)
**Purpose:** All exchange/price feed errors

**Categories:**
- Symbol validation errors
- API call errors
- Data parsing errors
- Format errors

**Common Error Codes:**
- `6001` - Symbol validation failed
- `6002` - API call failed
- `6003` - Data parsing failed
- `6004` - Rate limit exceeded
- `6005` - Invalid exchange response
- `6006` - Exchange unavailable
- `6007` - Authentication failed

---

### 7. ALEO CONTEXT ERRORS (7000-7999)
**Purpose:** Aleo context initialization errors

**Categories:**
- Context initialization errors
- Key generation errors
- Signature verification errors

**Common Error Codes:**
- `7001` - Context initialization failed
- `7002` - Key generation failed
- `7003` - Signature verification failed
- `7004` - Aleo network connection failed

---

## Usage Examples

### Creating an Error
```go
// Validation error
return appErrors.NewAppError(appErrors.ErrInvalidURL)

// Enclave error
return appErrors.NewAppError(appErrors.ErrReadingTargetInfo)

// Data extraction error
return appErrors.NewAppError(appErrors.ErrHTTPRequestFailed)
```

### Error Handling
```go
if err != nil {
    switch err.Code {
    case 1001:
        // Handle URL validation error
        log.Error("Invalid URL provided")
    case 2001:
        // Handle enclave error
        log.Error("Enclave quote generation failed")
    case 4001:
        // Handle HTTP request error
        log.Error("HTTP request failed")
    default:
        // Handle unknown error
        log.Error("Unknown error occurred")
    }
}
```