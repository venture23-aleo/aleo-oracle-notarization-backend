# Aleo Oracle Notarization Backend

## Overview

The Aleo Oracle Notarization Backend is a secure service that acts as a data oracle, running inside an Intel SGX enclave. The enclave is uniquely identified by a fingerprint called **MRENCLAVE**, which proves the code and configuration running inside. This backend fetches external data (such as cryptocurrency prices), processes it securely within the enclave, and generates a signed attestation report. This report contains both the unique enclave ID and the fetched oracle data, and can be verified by other systems—either by a separate backend or directly on-chain in an Aleo Oracle smart contract—to ensure the data is authentic and was produced by a trusted enclave.

The backend can handle different types of data, including integers, floats, and strings. Every time the backend restarts, the enclave automatically creates a new public-private key pair. The Aleo oracle contract keeps track of the current public key, so it can verify signatures from the enclave. This design guarantees that the oracle data is both trustworthy (integrity) and up-to-date (freshness), supporting decentralized and trust-minimized applications.

## 🚀 Quick Start

### Prerequisites
- **Docker** (with BuildKit enabled)
- **Intel SGX hardware** (with DCAP support and drivers installed)
- **Access to the SGX device files** (`/dev/sgx/enclave`, `/dev/sgx/provision`)

### Basic Setup
```sh
# Generate enclave signing key
make gen-key

# Build and run the application
make docker-build
make docker-run
```

### Common Commands
```sh
make build         # Build Go binary
make test          # Run tests
make docker-run    # Build and run with Docker
make help          # Show all available commands
```

## 📚 Documentation

For detailed documentation, see the [`docs/`](docs/) folder:

- **[Architecture & Working Flow](docs/architecture.md)** - Technical implementation details
- **[Setup & Installation Guide](docs/setup-guide.md)** - Complete setup instructions
- **[Makefile Guide](docs/makefile-guide.md)** - All available commands and usage

## 🔧 Configuration

The application contains config.json `internal/configs/config.json` files with default values. Overriding the default values will change the enclave measurement hash (MRENCLAVE).

- **PORT** - The port for the application.
- **WHITELISTED_DOMAINS** - The domains that are whitelisted for the attestation.
- **TCB_STATUS** - The status of the enclave's Trusted Computing Base (TCB) level.  This value is optional and determined during remote attestation by submitting the generated SGX quote to a verifier. It reflects a dynamic evaluation based on the platform’s CPU microcode, firmware version, and patch level. If not provided, the TCB status will be set to 5.
- **METRICS_PORT** - The port for the metrics server.
- **PRICE_FEED_CONFIG** - The configuration for the price feed.
- **LOG_LEVEL** - The level of logging for the application.
- **VERIFIER_ENDPOINT** - The endpoint for the verifier. This is used to get the TCB status of the enclave.

## 🛡️ Security

This application runs inside an Intel SGX enclave for enhanced security. The enclave provides:

- **Code integrity** - Ensures the application code hasn't been tampered with
- **Data confidentiality** - Protects sensitive data from the host system
- **Attestation** - Proves the application is running in a genuine enclave

## 🔄 Reproducibility

To reproduce the enclave measurement hash (MRENCLAVE), you can run the following commands:

```sh
make gen-key
make get-enclave-info
``` 

It will generate/update a `enclave_info.json` file in the root of the project. You can use this file to verify the enclave measurement hash (MRENCLAVE) of the running enclave. Ensure that the Intel SGX driver is loaded and the device files `sgx_enclave` and `sgx_provision` are accessible.  

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 🔗 Links

- [Aleo Oracle Documentation](https://aleo-oracle-docs.surge.sh/)
- [Gramine Framework](https://gramine.readthedocs.io/en/stable/)
- [Intel SGX](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html)

