# Docker Deployment Guide

This guide covers how to run the Aleo Oracle Notarization Backend using Docker containers in SGX supported machines, ensuring reproducible builds and consistent enclave measurements.

## Prerequisites

- **Linux distribution** (Ubuntu 20.04+ recommended for host)
- **Intel SGX hardware**
- **Access to SGX device files** (`/dev/sgx_enclave`, `/dev/sgx_provision`)
- **make** - make for automation
- **Docker** (with BuildKit enabled)
- **Docker Compose** for orchestration

### Required Tools
- `docker` - Docker engine
- `docker compose` - Docker Compose for orchestration
- `make` - Make utility for automation

## 🚀 Quick Start Options

We have **two approaches** to set up Docker environment and build and run the Docker image:

### Option 1: Automated Setup (Recommended)
**Use Make targets for quick, automated Docker deployment.** It will install Docker, generate the enclave signing key, generate the manifest template and build the Docker image.

```bash
# Install Docker components and setup the enclave signing key
make docker-setup

# Build and run the Docker container with SGX support
make docker-run
```

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

#### Verify that Docker is properly installed and configured:

```bash
# Check Docker version
docker --version

# Check Docker Compose version
docker compose version
```

**Expected Output:**
```
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
```
>> Generating enclave private key at deployments/secrets/enclave-signing-key.pem using OpenSSL ...
>> Setting permissions for deployment/secrets/enclave-signing-key.pem to 600 ...
>> Enclave signing key generated successfully!
-rw------- 1 user user 2484 Aug  4 05:50 deployment/secrets/enclave-signing-key.pem
```

### Step 3: Generate Manifest Template

Generate the manifest template for the Docker image.

```bash
# Generate manifest template
make generate-manifest-template
```

**Expected Output:**
```
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
```
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
```
>> Docker container status:
NAME                               IMAGE                                     COMMAND             SERVICE                            CREATED      STATUS          PORTS
aleo-oracle-notarization-backend   aleo-oracle-notarization-backend:latest   "./entrypoint.sh"   aleo-oracle-notarization-backend   2 days ago   Up 12 seconds   0.0.0.0:8000-8001->8000-8001/tcp, [::]:8000-8001->8000-8001/tcp
```


### Step 7: Optional - Setup Monitoring (Prometheus & Alertmanager)

For production deployments, you may want to set up monitoring and alerting. Please refer to the [Prometheus and Alertmanager Setup Guide](monitoring-setup-guide.md) for more details.

## Reproducible Builds

### ✅ Docker Deployment Advantages

**Docker deployment guarantees reproducible builds** because:

- **Pinned Gramine version** - Uses specific Gramine version (`gramineproject/gramine:stable-jammy@sha256:84b3d222e0bd9ab941f0078a462af0dbc5518156b99b147c10a7b83722ac0c38`)
- **Pinned Go version** - Uses specific Go version (`golang:1.24.4-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a`)
- **Static Configuration** - All config files included in build context
- **Deterministic Process** - Same build steps every time
- **Version Control** - All dependencies pinned to specific versions
- **No Host Dependencies** - Everything containerized except SGX devices

### Enclave Measurement Consistency
```bash
# Extract enclave measurement
make extract-enclave-artifacts
```

**Expected Output:**
```
Measurement: 211403836def8ebf5082b13681260ed4b6da6f26de3a17d041f8e7a388d67fb1
```

## Next Steps

- Read the [Architecture Guide](architecture.md) for technical details
- Explore the [API Documentation](api-documentation.md) for integration details
- For native deployment, see the [Native Deployment Guide](native-deployment-guide.md) 