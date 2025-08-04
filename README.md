# Aleo Oracle Notarization Backend

## Overview

The Aleo Oracle Notarization Backend is a secure service that acts as a data oracle, running inside an Intel SGX enclave. The enclave is uniquely identified by a fingerprint called **MRENCLAVE**, which proves the code and configuration running inside. This backend fetches external data (such as cryptocurrency prices), processes it securely within the enclave, and generates a signed attestation report. This report contains both the unique enclave ID and the fetched oracle data, and can be verified by other systems—either by a separate backend or directly on-chain in an Aleo Oracle smart contract—to ensure the data is authentic and was produced by a trusted enclave.

The backend can handle different types of data, including integers, floats, and strings. Every time the backend restarts, the enclave automatically creates a new public-private key pair. The Aleo oracle contract keeps track of the current public key, so it can verify signatures from the enclave. This design guarantees that the oracle data is both trustworthy (integrity) and up-to-date (freshness), supporting decentralized and trust-minimized applications.

## 📚 Documentation

For detailed documentation, see the [`docs/`](docs/) folder:

- **[Architecture & Working Flow](docs/architecture.md)** - Technical implementation details
- **[Deployment Guide](docs/deployment-guide.md)** - Complete deployment instructions
- **[Docker Deployment Guide](docs/docker-deployment-guide.md)** - Complete docker deployment instructions
- **[Native Deployment Guide](docs/native-deployment-guide.md)** - Complete native deployment instructions
- **[API Documentation](docs/api-documentation.md)** - Complete API reference with endpoints, examples, and usage guidelines
- **[Error Codes Reference](docs/error-codes.md)** - Complete reference for all error codes, troubleshooting, and usage examples

## 🔧 Configuration

The application contains config.json `internal/configs/config.json` files with default values. Overriding the default values will change the enclave measurement hash (MRENCLAVE) so please be careful when overriding the values.

- **PORT** - The port for the application.
- **WHITELISTED_DOMAINS** - The domains that are whitelisted for the attestation.
- **METRICS_PORT** - The port for the metrics server.
- **PRICE_FEED_CONFIG** - The configuration for the price feed.
- **LOG_LEVEL** - The level of logging for the application.

## 🛡️ Security

This application runs inside an Intel SGX enclave for enhanced security. The enclave provides:

- **Code integrity** - Ensures the application code hasn't been tampered with
- **Data confidentiality** - Protects sensitive data from the host system
- **Attestation** - Proves the application is running in a genuine enclave

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

