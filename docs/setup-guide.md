# Setup & Installation Guide

## Prerequisites

- **Docker** (with BuildKit enabled)
- **Intel SGX hardware** (with DCAP support and drivers installed)
- **Access to the SGX device files** (`/dev/sgx/enclave`, `/dev/sgx/provision`)
- **A Gramine-compatible SGX signing key** (e.g., `enclave-key.pem`)
- **Docker Compose** for easier orchestration

## 1. Prepare Your Environment

- Ensure your user has permission to run Docker commands.
- Make sure your SGX device files are present on the host:
  ```sh
  ls /dev/sgx/enclave /dev/sgx/provision
  ```

- **Generating the Enclave Private Key:**

  Before building the Docker image, you need an enclave signing key (private key) to sign the Gramine manifest. This key must be an RSA 3072-bit key with public exponent 3, as required by Intel SGX.

  ```sh
  make gen-key         # Uses gramine-sgx-gen-private-key (requires Gramine)
  make gen-key-openssl # Uses OpenSSL with exponent 3 (requires OpenSSL 1.1.1+)
  ```
    Both commands will create `secrets/enclave-key.pem` suitable for SGX signing.

## 2. Build and Run the Docker Image

All Docker Compose and Docker commands are wrapped by the Makefile for simplicity and consistency. **Please use the provided Makefile targets for building and running the application:**

```sh
make docker-build  # Build the Docker image (with manifest)
make docker-run    # Build and run the container using Docker Compose
```

This ensures all arguments, secrets, and environment variables are set correctly. You do not need to run `docker compose build` or `docker run` directly.

## 3. Configuration Files

- **Static `/etc/hosts` and `/etc/resolv.conf`:**  
  The container uses static versions for enclave isolation.  
  - `build/inputs/static_hosts` (should contain at least `127.0.0.1 localhost`)
  - `build/inputs/static_resolv.conf` (should contain valid DNS servers, e.g., `nameserver 8.8.8.8`)

- **SGX Quote Configuration:**  
  - `build/inputs/sgx_default_qcnl.conf` should point to a reachable PCCS server for DCAP attestation.

## 4. Troubleshooting

- **SGX errors:** Ensure your host has the correct drivers and device files, and that they are mapped into the container.
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