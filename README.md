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

### mTLS Reverse Proxy (Nginx)

For production / secure environments, an optional Nginx reverse proxy is provided (Docker-based) that:

- Terminates TLS and enforces **mutual TLS (mTLS)** — every client must present a certificate signed by your internal Root CA.
- Proxies validated requests to the internal notarization backend over the Docker network (plain HTTP).
- Protects metrics and API surface (only accessible via authenticated mTLS channel).

Artifacts are generated via `make generate-mtls-certs` (automatically included in `make docker-setup`). The unified script `deployment/secrets/generate-mtls-certs.sh` also manages client cert lifecycle (generate, renew, revoke, list, show). Files produced under `deployment/secrets/mtls/`:

- `ca.crt` / `ca.key` – Root Certificate Authority
- `server.crt` / `server.key` – Nginx server certificate (SAN includes `localhost`, `nginx-mtls-proxy`, and backend service name)
- `client.crt` / `client.key` – Sample client identity certificate

Default subject values:

- Country (C): NP
- Organization (O): Venture23

Override per client with: `make client-cert-generate CN=alice COUNTRY=US ORG=Research`.

Example secure request (default generated client `AleoOracleClient`):

```bash
curl --cacert deployment/secrets/mtls/ca.crt \
  --cert deployment/secrets/mtls/client.crt \
  --key deployment/secrets/mtls/client.key \
  https://localhost:8443/health
```

You can distribute only `ca.crt` to trusted clients so they can verify your service; each client should have a unique certificate.

Manage client certificates with Make targets (examples):

```bash
make client-cert-generate CN=alice DAYS=120
make client-cert-renew CN=alice DAYS=365
make client-cert-show CN=alice
make client-cert-list
make client-cert-revoke CN=alice
```

Advanced (custom subject / SAN):

```bash
make client-cert-generate CN=service-a ORG=MyOrg COUNTRY=DE SAN=URI:spiffe://oracle/service/service-a
make client-cert-generate CN=web-user SAN=DNS:user.example.internal
make client-cert-renew CN=service-a DAYS=400 SAN=URI:spiffe://oracle/service/service-a,DNS:svc-a.internal
```

### Rate Limiting & Tuning

Basic rate limiting and connection limiting have been added with sensible defaults intended to protect the backend from accidental floods while allowing normal operation.

#### Managing Per-Client Certificates

Additional client certificates can be managed via the helper script:

```bash
# Generate (CN required). Optional: DAYS=365 make client-cert-generate CN=alice
make client-cert-generate CN=alice

# List issued client certs
make client-cert-list

# Show certificate details
make client-cert-show CN=alice

# Renew (re-issue) certificate with new validity
make client-cert-renew CN=alice DAYS=400

# Revoke (local record only – implement CRL/OCSP for enforcement)
make client-cert-revoke CN=alice
```

Generated client certs are stored under `deployment/secrets/mtls/clients/<cn>/`.

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
