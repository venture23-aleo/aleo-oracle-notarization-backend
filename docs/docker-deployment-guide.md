# Docker Deployment Guide

This guide covers how to run the Aleo Oracle Notarization Backend using Docker containers in SGX supported machines, ensuring reproducible builds and consistent enclave measurements.

## Prerequisites

- **Linux distribution** (Ubuntu 20.04+ recommended for host)
- **Intel SGX hardware**
- **Access to SGX device files** (`/dev/sgx_enclave`, `/dev/sgx_provision`)
- **make** - make for automation
- **Docker** (with BuildKit enabled)

## 🚀 Quick Start Options

### Option 1: Automated Setup (Recommended)

**Use Make targets for quick, automated Docker deployment.** It will install Docker, generate the enclave signing key, generate the manifest template and build the Docker image.

```bash
# Install Docker components and setup the enclave signing key
make docker-setup

# Build and run the Docker container with SGX support
make docker-run
```

If you want the **secure mTLS reverse proxy**, it is already integrated. The `nginx` service exposes port `8443` with mandatory client certificate verification.

### Option 2: Step-by-Step Manual Setup

**Follow the detailed step-by-step process below for more control and understanding.**
Please refer to the [Makefile](../deployment/docker/Makefile) for the full list of targets.

**Note:** Each `make` target below may depend on previous targets, so running them individually can result in redundant steps. For daily use, prefer the automated approach above.

### Step 1: Install Docker and Docker Compose

First, install Docker and Docker Compose on your system and verify the installation:

#### Option A: Using Make Target (Recommended)

```bash
# Complete Docker setup
make docker-install
```

#### Option B: Manual Installation

For manual installation, refer to the [official Docker installation guide](https://docs.docker.com/engine/install/) for your specific Linux distribution.

#### Verify that Docker is properly installed and configured

```bash
# Check Docker version
docker --version

# Check Docker Compose version
docker compose version
```

**Expected Output:**

```text
Docker version 28.3.3, build 980b85
Docker Compose version v2.37.3
```

### Step 2: Generate Enclave Signing Key

Generate the RSA 3072-bit private key required for SGX enclave signing:

```bash
# Generate enclave signing key
make generate-enclave-signing-key

# Verify the key was created
ls -la deployment/secrets/enclave-signing-key.pem
```

**Expected Output:**

```text
>> Generating enclave private key at deployments/secrets/enclave-signing-key.pem using OpenSSL ...
>> Setting permissions for deployment/secrets/enclave-signing-key.pem to 600 ...
>> Enclave signing key generated successfully!
-rw------- 1 user user 2484 Aug  4 05:50 deployment/secrets/enclave-signing-key.pem
```

### Step 3: Generate Manifest Template

```bash
# Generate manifest template
make generate-manifest-template
```

**Expected Output:**

```text
>> Gramine manifest template generated as deployment/docker/inputs/aleo-oracle-notarization-backend.manifest.template
```

### Step 4: Build Docker Image

First, it will generate the manifest template and then build the Docker image.

```bash
# Generate manifest template and build the Docker image
make docker-build

# Verify the image was created
docker images | grep aleo-oracle-notarization-backend
```

### Step 5: Run the Docker Container

Start the application using Docker Compose:

```bash
# Run the container
make docker-run

```

### Step 6: Verify Application is Running

Check that the application is running correctly:

```bash
# View container logs
make docker-logs
```

**Expected Output:**

```text
time=2025-07-31T10:02:35.983Z level=INFO msg="Configuration validation passed"
time=2025-07-31T10:02:36.450Z level=INFO msg="Starting system metrics collector"
time=2025-07-31T10:02:36.451Z level=INFO msg="Notarization server started" address=:8000
time=2025-07-31T10:02:36.452Z level=INFO msg="Metrics server started" address=:8001
```

```bash
# Check container status
make docker-status
```

**Expected Output:**

```text
>> Docker container status:
NAME                               IMAGE                                     COMMAND             SERVICE                            CREATED      STATUS          PORTS
nginx-mtls-proxy                   aleo-oracle-notarization-backend-nginx    "/entrypoint.sh"    nginx                              2 days ago   Up 12 seconds   0.0.0.0:8443->8443/tcp
aleo-oracle-notarization-backend   aleo-oracle-notarization-backend:latest   "./entrypoint.sh"   aleo-oracle-notarization-backend   2 days ago   Up 12 seconds   (internal: 8000,8001)
```

### Step 7: Optional - Setup Monitoring (Prometheus & Alertmanager)

For production deployments, you may want to set up monitoring and alerting. Please refer to the [Prometheus and Alertmanager Setup Guide](monitoring-setup-guide.md) for more details.

## Reproducible Builds

### ✅ Docker Deployment Advantages

**Docker deployment guarantees reproducible builds** because:

- **Pinned Gramine version** - Uses specific Gramine version (`gramineproject/gramine:stable-jammy@sha256:84b3d222e0bd9ab941f0078a462af0dbc5518156b99b147c10a7b83722ac0c38`)
- **Pinned Go version** - Uses specific Go version (`golang:1.24.4-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a`)
- **Static Configuration** - All config files included in build context

```bash
# Extract enclave measurement
make extract-enclave-artifacts
```

**Expected Output:**

```text
Measurement: 211403836def8ebf5082b13681260ed4b6da6f26de3a17d041f8e7a388d67fb1
```

## 🔐 mTLS Reverse Proxy (Nginx)

An optional (enabled by default in Compose) Nginx layer terminates TLS and enforces client certificate authentication. This design ensures only authenticated clients reach the enclave-backed service.

### Generated Certificates

Certificates are generated by:

```bash
make generate-mtls-certs
```

Artifacts (ignored by git) are stored in `deployment/secrets/mtls/`:

| File | Purpose |
|------|---------|
| `ca.crt` | Root CA certificate (distribute to clients) |
| `ca.key` | Root CA private key (keep secret) |
| `server.crt` / `server.key` | Nginx server identity |
| `client.crt` / `client.key` | Example client certificate |

Default subject values:

| Field | Default |
|-------|---------|
| Country (C) | NP |
| Organization (O) | Venture23 |

Override when generating a client cert, e.g.:

```bash
make client-cert-generate CN=team1 COUNTRY=US ORG=Analytics
```

### Test mTLS Endpoint

```bash
curl --cacert deployment/secrets/mtls/ca.crt \
  --cert deployment/secrets/mtls/client.crt \
  --key deployment/secrets/mtls/client.key \
  https://localhost:8443/health
```

If the client certificate is omitted or invalid, Nginx returns `400` or `495/496` (client certificate errors).

### Rotating Certificates

To rotate certificates, remove the `deployment/secrets/mtls/` directory and rerun:

```bash
rm -rf deployment/secrets/mtls
make generate-mtls-certs
make docker-run
```

Distribute the updated `ca.crt` and new client certificates as needed.

### Managing Client Certificates

The unified script `deployment/secrets/generate-mtls-certs.sh` supports client operations via Make:

```bash
make client-cert-generate CN=team1
make client-cert-renew CN=team1 DAYS=300
make client-cert-list
make client-cert-show CN=team1
make client-cert-revoke CN=team1
```

Advanced examples (custom org / country / SAN entries):

```bash
make client-cert-generate CN=svc-api ORG=CoreServices COUNTRY=US SAN=URI:spiffe://oracle/client/svc-api
make client-cert-generate CN=analytics COUNTRY=NL SAN=DNS:analytics.internal.example.com
make client-cert-renew CN=svc-api DAYS=365 SAN=URI:spiffe://oracle/client/svc-api,DNS:svc-api.internal
```

Direct script usage:

```bash
deployment/secrets/generate-mtls-certs.sh generate-client --cn team1 --days 180
deployment/secrets/generate-mtls-certs.sh renew-client --cn team1 --days 365
deployment/secrets/generate-mtls-certs.sh list-clients
deployment/secrets/generate-mtls-certs.sh show-client --cn team1
deployment/secrets/generate-mtls-certs.sh revoke-client --cn team1
```

Revocations are recorded locally (fingerprint stored under `mtls/revoked/`). To enforce revocation at the proxy, integrate CRL/OCSP and reload Nginx.

### Disabling Nginx Layer

If you need to expose the backend directly (NOT recommended for production), override the services:

```bash
DOCKER_SERVICES=aleo-oracle-notarization-backend make docker-run
```

Or comment/remove the `nginx` service in `docker-compose.yml`.

### Managing Additional Client Certificates

You can create, renew, list, show, and revoke individual client identities without regenerating the Root CA:

```bash
# Generate a new client cert (Common Name required)
make client-cert-generate CN=alice

# List all issued certs (CN + fingerprint)
make client-cert-list

# Show full certificate details
make client-cert-show CN=alice

# Renew (re-issue) an existing cert for a different validity window
make client-cert-renew CN=alice DAYS=365

# Generate with SAN (multiple SANs comma-separated or use multiple SAN= invocations via direct script)
make client-cert-generate CN=alice SAN=DNS:alice.internal.example.com

# Revoke (records fingerprint locally; enforcement requires CRL / OCSP integration)
make client-cert-revoke CN=alice
```

Artifacts live under `deployment/secrets/mtls/clients/<cn>/`.

## Next Steps

- Read the [Architecture Guide](architecture.md) for technical details
- Explore the [API Documentation](api-documentation.md) for integration details
- For native deployment, see the [Native Deployment Guide](native-deployment-guide.md)
