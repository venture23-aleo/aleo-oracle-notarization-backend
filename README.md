# Aleo Oracle Notarization Backend

## Overview

The Aleo Oracle Notarization Backend is a secure service that acts as a data oracle, running inside an Intel SGX enclave. The enclave is uniquely identified by a fingerprint called **MRENCLAVE**, which proves the code and configuration running inside. This backend fetches external data (such as cryptocurrency prices), processes it securely within the enclave, and generates a signed attestation report. This report contains both the unique enclave ID and the fetched oracle data, and can be verified by other systems—either by a separate backend or directly on-chain in an Aleo Oracle smart contract—to ensure the data is authentic and was produced by a trusted enclave.

The backend can handle different types of data, including integers, floats, and strings. Every time the backend restarts, the enclave automatically creates a new public-private key pair. The Aleo oracle contract keeps track of the current public key, so it can verify signatures from the enclave. This design guarantees that the oracle data is both trustworthy (integrity) and up-to-date (freshness), supporting decentralized and trust-minimized applications.

---

## Notarization Backend Working Flow

### 1. Initialization
- When the backend starts, it generates an Aleo private key.
- This key is subsequently used to sign the complete attestation report after it has been serialized and encoded in an Aleo-specific format.

### 2. Attestation Request
- A user submits an attestation request to the backend.
- The backend performs the following steps:
  - Checks whether the target URL is included in the whitelist.
  - Validates the format of the attestation request payload.
  - If the request is valid, it fetches the data that needs to be attested.

### 3. Oracle Proof Data Preparation
- The backend constructs the oracle report data, which includes:
  - Details of the attestation request,
  - A timestamp,
  - The attestation data itself.

### 4. Hashing and Encoding
- **Encoded Request:**  
  To generate the encoded request, the attestation data and timestamp fields in the report data are temporarily set to zero. This standardizes the report data for hashing, ensuring the hash does not include the actual attestation data or timestamp.
- **Request Hash:**  
  A Poseidon8 hash is computed from the encoded request, producing a unique identifier for the request parameters (excluding attestation data and timestamp).
- **Timestamped Request Hash:**  
  The request hash is combined with the actual timestamp, and this combination is hashed again using Poseidon8. This step binds the request to a specific point in time.
- **Attestation Hash:**  
  The complete report data, now including the real attestation data and timestamp, is hashed using Poseidon8. The resulting attestation hash is embedded within the SGX quote to prove the integrity and origin of the attested data.

### 5. Quote Generation
- **Quoting Enclave (QE):**  
  Generates a quote, which is a signed statement containing enclave details, the attestation hash, a signature created with the PCK private key, and the PCK certification chain.
- **Provisioning Enclave (PE):**  
  Assists in provisioning cryptographic keys during the initialization of the QE.

### 6. Final Signing and Output
- The oracle report is generated from the quote.
- A Poseidon8 hash of the oracle report is signed using the Aleo private key.
- The backend sends:
  - The attestation report to the remote verifier backend.
  - The oracle data to the Aleo oracle program.

### 7. Verification
- The verifier backend and the Aleo program check that:
  - The quote originates from a genuine enclave.
  - The data is correctly signed and has not been tampered with.

## Gramine Framework

This project uses the [Gramine Framework](https://gramine.readthedocs.io/en/stable/) to run the notarization backend inside an Intel SGX enclave. Gramine is a lightweight guest OS designed to run a single Linux application with minimal host requirements, providing:

- **SGX compatibility:** Seamless execution of unmodified Linux applications inside Intel SGX enclaves.
- **Security:** Additional protection against side-channel attacks and strong isolation from the host and container environment.
- **Manifest-based configuration:** Fine-grained control over file system access, environment variables, and resource limits.
- **File system isolation:** Ability to provide static `/etc/hosts` and `/etc/resolv.conf` to the enclave, ensuring deterministic and secure runtime behavior.

**Note:**  
You do not need to install Gramine on your host unless you want to generate the enclave signing key using `gramine-sgx-gen-private-key` or run/debug Gramine applications outside Docker. All necessary Gramine tools and runtime are included in the Docker image.

For more details, see the [Gramine Documentation](https://gramine.readthedocs.io/en/stable/).

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
| PCK Certificate Chain	|Public cert chain starting from the PCK cert up to Intel's root CA.

### 4. **QE Report + QE Report Signature**
To prove that the Quoting Enclave itself is trustworthy.

| Field	| Description|
| --------- | ----------- |
| QE Report	|A report generated by the QE about itself.
| QE Report Signature	|Signed by the platform's quoting infrastructure (used by the verifier to validate the QE).

---


## Using the Makefile for Common Tasks

A `Makefile` is provided to simplify the most common development, build, and deployment tasks. The Makefile will automatically use variables from your `.env` file if present, so you can configure your environment in one place.

### Common Make Targets

- `make build`         – Build the Go binary for your application.
- `make run`           – Build and run the application locally.
- `make test`          – Run unit tests.
- `make docker-build`  – Build the Docker image (including manifest generation).
- `make docker-run`    – Build the Docker image and run the container using Docker Compose.
- `make gen-key`         – Generate the enclave private key for SGX signing (Gramine tool).
- `make gen-key-openssl` – Generate the enclave private key for SGX signing (OpenSSL).
- `make clean`         – Remove built binaries.
- `make help`          – Show a summary of available make targets.

### Example Usage

```sh
make build         # Compile the Go binary
make run           # Build and run locally
make test          # Run tests
make docker-build  # Build Docker image (with manifest)
make docker-run    # Build and run with Docker Compose
make gen-key         # Generate the enclave private key for SGX signing (Gramine tool)
make gen-key-openssl # Generate the enclave private key for SGX signing (OpenSSL)
make clean         # Clean up binaries
make help          # Show help
```

> **Tip:** The Makefile will use variables from your `.env` file (such as `PORT` and `WHITELISTED_DOMAINS`) if present, so you can easily configure your build and run environment.

> **Note:** The Makefile will automatically configure Docker Compose with the correct application name, private key, and port mapping based on its targets, so you do not need to manually edit `docker-compose.yml` for these values.


## Getting Started: Building and Running the Application

### Prerequisites

- **Docker** (with BuildKit enabled)
- **Intel SGX hardware** (with DCAP support and drivers installed)
- **Access to the SGX device files** (`/dev/sgx/enclave`, `/dev/sgx/provision`)
- **A Gramine-compatible SGX signing key** (e.g., `enclave-key.pem`)
- (Optional) **Docker Compose** for easier orchestration

### 1. Prepare Your Environment

- Ensure your user has permission to run Docker commands.
- Make sure your SGX device files are present on the host:
  ```sh
  ls /dev/sgx/enclave /dev/sgx/provision
  ```
- Place your SGX signing key (e.g., `enclave-key.pem`) in a secure location on your host. It is recommended to store it in the `secrets` folder at the root of this project.

- **Setup `.env` file:**

  The application uses the `WHITELISTED_DOMAINS` environment variable to control which external domains are allowed for data fetching. This is a security feature to ensure only trusted sources are queried by the backend.
  
  1. Copy the provided `.env_sample` file to `.env` in the project root:

      ```sh
      cp .env_sample .env
      ```
  2. Edit the `WHITELISTED_DOMAINS` line in your `.env` file to include all domains you want to allow, separated by commas:
      ```env
      WHITELISTED_DOMAINS=example.com,api.example.com,another.com
      ```
  3. Set the `PORT` variable as needed (default is 8080):
      ```env
      PORT=8080
      ```

    **Note:** Only `PORT` and `WHITELISTED_DOMAINS` are required in your `.env` file. Do not include secrets or sensitive information.

    Refer to `.env_sample` for an example configuration.

- **Generating the Enclave Private Key:**

  Before building the Docker image, you need an enclave signing key (private key) to sign the Gramine manifest. This key must be an RSA 3072-bit key with public exponent 3, as required by Intel SGX.

  ```sh
  make gen-key         # Uses gramine-sgx-gen-private-key (requires Gramine)
  make gen-key-openssl # Uses OpenSSL with exponent 3 (requires OpenSSL 1.1.1+)
  ```
    Both commands will create `secrets/enclave-key.pem` suitable for SGX signing.

### 2. Build and Run the Docker Image

All Docker Compose and Docker commands are wrapped by the Makefile for simplicity and consistency. **Please use the provided Makefile targets for building and running the application:**

```sh
make docker-build  # Build the Docker image (with manifest)
make docker-run    # Build and run the container using Docker Compose
```

This ensures all arguments, secrets, and environment variables are set correctly. You do not need to run `docker compose build` or `docker run` directly.

### 3. Configuration Files

- **Static `/etc/hosts` and `/etc/resolv.conf`:**  
  The container uses static versions for enclave isolation.  
  - `build/inputs/static_hosts` (should contain at least `127.0.0.1 localhost`)
  - `build/inputs/static_resolv.conf` (should contain valid DNS servers, e.g., `nameserver 8.8.8.8`)

- **SGX Quote Configuration:**  
  - `build/inputs/sgx_default_qcnl.conf` should point to a reachable PCCS server for DCAP attestation.

### 4. Troubleshooting

- **SGX errors:** Ensure your host has the correct drivers and device files, and that they are mapped into the container.
- **Network errors:** Ensure your static `resolv.conf` contains valid DNS servers.
- **Manifest errors:** Check for syntax issues in your Gramine manifest template.

For more details, see the comments in the Dockerfile and the Gramine documentation: https://gramine.readthedocs.io/en/stable/

## Reproducible Builds

This project is designed to support reproducible builds, ensuring that the same source and configuration will always produce the same output binaries and container images. Here's how reproducibility is achieved:

- **Pinned Base Images and Package Versions:**
  - The Dockerfile uses a specific, versioned base image (e.g., `gramineproject/gramine:stable-jammy`).
  - All package installations are performed in a controlled, minimal environment to avoid unexpected updates.

- **Static Configuration Files:**
  - Files such as `/etc/hosts`, `/etc/resolv.conf`, `/etc/sgx_default_qcnl.conf` and manifest templates are included in the build context and copied into the image, ensuring consistent configuration across builds.

- **Deterministic Key Generation:**
  - The enclave signing key is generated using a standard process (`make gen-key` or `make gen-key-openssl`), and you can check the key's properties for consistency.

- **Standardized Build Process:**
  - All build and run steps are wrapped in Makefile targets, so every user and CI system runs the same commands with the same environment variables and arguments.

- **Environment Variables:**
  - Only non-sensitive, required variables are used from `.env`, and these are documented and versioned for consistency.

- **No Host Dependencies:**
  - The build does not depend on host-specific files or settings, except for the required SGX devices and the enclave key, which are explicitly managed.

- **Consistent Enclave Identity (MRENCLAVE):**
  - Because the build process is reproducible and all configuration, code, and dependencies are fixed, the enclave measurement (MRENCLAVE) will remain the same for identical builds. This ensures that the enclave identity is consistent across deployments, which is critical for attestation and trust in SGX-based systems.


