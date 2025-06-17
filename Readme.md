# Aleo Oracle Notarization Backend

## Overview

The notarization backend is a secure data oracle service that runs inside an Intel SGX enclave, identified by a unique fingerprint called **MRENCLAVE**. It fetches external data (like crypto prices), securely processes it within the enclave, and produces a signed attestation report that includes the enclave ID and the oracle data. This report can be verified by a separate backend and directly on-chain in an Aleo Oracle program to confirm its authenticity.

The backend supports multiple data encoding formats such as integer, float, and string. Additionally, on every restart, the enclave generates a new public-private key pair, and the Aleo oracle contract keeps track of the current public key for verification purposes. This architecture ensures both the integrity and freshness of the oracle data in a decentralized, trust-minimized environment.

---

## Notarization Backend Working Flow

### 1. Initialization
- On startup, the backend generates an Aleo private key.
- This key is later used to sign the full attestation report after serialization and Aleo-specific encoding.

### 2. Attestation Request
- A user sends an attestation request to the backend.
- The backend:
  - Checks if the target URL is whitelisted
  - Validates the attestation request payload format.
  - If valid, it fetches the data to be attested.

### 3. Oracle Proof Data Preparation
- Backend prepares the oracle report data, embedding:
  - Attestation request details
  - Timestamp
  - Attestation data

### 4. Hashing and Encoding
- **Encoded Request:**
  - The attestation data and timestamp fields in the report data are temporarily set to 0.
  - This version of the report data is called the encoded request.
- **Request Hash:**
  - A Poseidon8 hash is computed from the encoded request.
- **Timestamped Request Hash:**
  - Combines the request hash with timestamp and hashes again using Poseidon8.
- **Attestation Hash:**
  - The full report data is hashed using Poseidon8.
  - This attestation hash is embedded inside the quote.

### 5. Quote Generation
- The **Quoting Enclave (QE):**
  - Generates a quote, a signed statement that includes enclave details, attestation hash, signature signed by the PCK private key, and PCK certification chain.
- The **Provisioning Enclave (PE):**
  - Assists in provisioning cryptographic keys during initialization of the QE.

### 6. Final Signing and Output
- The Oracle report is created from the quote.
- A Poseidon8 hash of the oracle report is signed using the Aleo private key.
- The backend sends:
  - Attestation report to the remote verifier backend.
  - Oracle data to the Aleo oracle program.

### 7. Verification
- The verifier backend and Aleo program confirm:
  - The quote is from a genuine enclave.
  - The data is signed correctly and not tampered with.

## Gramine Framework
We use Gramine Framework to run the notarization backend inside an Intel SGX enclave. It is a lightweight guest OS that is designed to run a single Linux application with minimal host requirements and provides additional protection against side-channel attacks. It wraps the application binary to make it compatible with the SGX environment. Please refer to the [Gramine Documentation](https://gramine.readthedocs.io/en/stable/) for more details.

## DCAP based Remote Attestation Flow

The DCAP attestation flow is as follows:

![Aleo Oracle DCAP Attestation Flow](https://gramine.readthedocs.io/en/stable/_images/dcap.svg "Aleo Oracle DCAP Attestation Flow")

## Contents of a SGX Quote (at a high level)
### 1. **Report Body (from the Application Enclave)**
This is the core enclave information, created by the target (application) enclave, and includes:

| Field	    | Description |
| --------- | ----------- |
|MRENCLAVE	|Measurement (SHA256 hash) of the enclave code and initial state.
|MRSIGNER	|Measurement of the entity that signed the enclave.
|ISV_SVN	    |Enclave's version number (Security Version Number).
|ATTRIBUTES	|Enclave configuration flags (e.g., debug, 64-bit, etc.).
|REPORTDATA	|64-byte field optionally filled by the enclave (e.g., to hash app-specific input or nonce).

### 2. **Quote Header**
Metadata about how the quote was created.

| Field	| Description|
| --------- | ----------- |
| Version	|Quote format version.
| Attestation Key Type	|Usually EPID (older) or ECDSA (modern).
| QE SVN	|Quoting Enclave Security Version Number.
| QE Vendor ID	|Intel vendor ID.
| User Data	|Optional app-specific data.

### 3. **Signature and Authentication**
Used to authenticate the quote to the verifier.

| Field	| Description|
| --------- | ----------- |
| Signature	|Signature over the report using the PCK private key.
| Auth Data	|Authenticated data (e.g., signature structure).
| PCK Certificate Chain	|Public cert chain starting from the PCK cert up to Intel’s root CA.

### 4. **QE Report + QE Report Signature**
To prove that the Quoting Enclave itself is trustworthy.

| Field	| Description|
| --------- | ----------- |
| QE Report	|A report generated by the QE about itself.
| QE Report Signature	|Signed by the platform’s quoting infrastructure (used by the verifier to validate the QE).
