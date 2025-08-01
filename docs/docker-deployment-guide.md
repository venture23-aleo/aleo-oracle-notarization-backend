# Docker Deployment Guide

This guide covers how to run the Aleo Oracle Notarization Backend using Docker containers with SGX support, ensuring reproducible builds and consistent enclave measurements.

## Prerequisites

### System Requirements
- **Linux distribution** (Ubuntu 20.04+ recommended for host)
- **Docker** (with BuildKit enabled)
- **Docker Compose** for orchestration
- **Intel SGX hardware** (with DCAP support and drivers installed)
- **Access to SGX device files** (`/dev/sgx_enclave`, `/dev/sgx_provision`)

### Required Tools
- `docker` - Docker engine
- `docker-compose` - Docker Compose for orchestration
- `make` - Make utility for automation

## 🚀 Quick Start Options

We have **two approaches** to set up the Docker environment:

### Option 1: Automated Setup (Recommended)
**Use Make targets for quick, automated Docker deployment.**
```bash
# Install Docker components and setup the enclave signing key
make docker-setup

# Build and run the Docker container with SGX support
make docker-run
```

### Option 2: Step-by-Step Manual Setup
**Follow the detailed step-by-step process below for more control and understanding.**


### Step 1: Install Docker and Docker Compose

First, install Docker and Docker Compose on your system:

#### Option A: Using Make Target (Recommended)
```bash
# Complete Docker setup
make docker-install
```

#### Option B: Manual Installation
For manual installation, refer to the [official Docker installation guide](https://docs.docker.com/engine/install/) for your specific Linux distribution.

### Step 2: Verify Docker Installation

Verify that Docker is properly installed and configured:

```bash
# Check Docker version
docker --version

# Check Docker Compose version
docker-compose --version

# Test Docker installation
docker run hello-world
```

**Expected Output:**
```
Docker version 28.3.0, build 38b7060
Docker Compose version v2.37.3
Hello from Docker!
This message shows that your installation appears to be working correctly.
```

### Step 3: Check SGX Hardware Support

Verify that your system has SGX hardware support:

```bash
# Check CPU SGX support
sudo cat /proc/cpuinfo | grep -i sgx

# Check if SGX device files exist
ls -la /dev/sgx_enclave 2>/dev/null || echo "SGX device files not found"

```

**Expected Output:**
```
# CPU check should show: sgx sgx_lc (in CPU flags)
# Device files should show: crw-rw-rw- 1 root sgx 10, 125 /dev/sgx_enclave
```

### Step 4: Generate Enclave Signing Key

Generate the RSA 3072-bit private key required for SGX enclave signing:

```bash
# Generate enclave signing key
make generate-enclave-signing-key

# Verify the key was created
ls -la secrets/enclave-signing-key.pem
```

**Expected Output:**
```
>> Generating enclave private key at secrets/enclave-signing-key.pem using OpenSSL ...
>> Enclave signing key generated successfully!
-rw------- 1 user user 1679 Jul 31 10:00 secrets/enclave-signing-key.pem
```

### Step 5: Build Docker Image

Build the Docker image with SGX support:

```bash
# Build the Docker image
make docker-build

# Verify the image was created
docker images | grep aleo-oracle-notarization-backend
```

**Expected Output:**
```
>> Building Docker image...
DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker compose build aleo-oracle-notarization-backend
aleo-oracle-notarization-backend   latest    abc123def456   2 minutes ago   1.2GB
```

### Step 6: Run the Docker Container

Start the application using Docker Compose:

```bash
# Run the container
make docker-run

# Check container status
make docker-status
```

**Expected Output:**
```
>> Running Docker container...
[+] Running 1/1
 ✔ Container aleo-oracle-sgx-vm-test-aleo-oracle-notarization-backend-1  Started
```

### Step 7: Verify Application is Running

Check that the application is running correctly:

```bash
# View container logs
make docker-logs

# Check container status
make docker-status
```

**Expected Output:**
```
time=2025-07-31T10:02:35.983Z level=INFO msg="Configuration validation passed"
time=2025-07-31T10:02:36.450Z level=INFO msg="Starting system metrics collector"
time=2025-07-31T10:02:36.451Z level=INFO msg="Notarization server started" address=:8000
time=2025-07-31T10:02:36.452Z level=INFO msg="Metrics server started" address=:8001
```

### Step 8: Optional - Setup Monitoring (Prometheus & Alertmanager)

For production deployments, you may want to set up monitoring and alerting:

#### Option A: Using Make Targets (Recommended)
```bash
# Setup complete monitoring stack
make setup-monitoring

# Or setup individual components
make setup-prometheus
make setup-alertmanager
```

#### Option B: Manual Setup
For manual setup, refer to the [Prometheus installation guide](https://prometheus.io/docs/prometheus/latest/getting_started/) and [Alertmanager documentation](https://prometheus.io/docs/alerting/latest/alertmanager/).


### Docker Compose Configuration
The application uses `docker-compose.yml` for orchestration:

```yaml
services:
  aleo-oracle-notarization-backend:
    image: ${APP}:latest
    container_name: notarization-backend
    build: 
      context: .
      dockerfile: ./docker/Dockerfile
      secrets:
        - gramine-private-key
      args:
        APP: ${APP}
    platform: linux/amd64
    ports:
      - 8000:8000
      - 8001:8001
    devices:
      - /dev/sgx_enclave:/dev/sgx_enclave
      - /dev/sgx_provision:/dev/sgx_provision
      
secrets:
  gramine-private-key:
    file: ./secrets/enclave-signing-key.pem
```

## Development Workflow

### Using Make Commands (Recommended)
The Makefile provides convenient commands for Docker development:

```bash
# Build and run
make docker-run

# Build only
make docker-build

# Run in foreground mode
make docker-run-fg

# Stop containers
make docker-stop

# View logs
make docker-logs

# Check status
make docker-status
```

### Using Docker Commands Directly
You can also use Docker commands directly:

```bash
# Build image
docker compose build

# Run container
docker compose up -d

# View logs
docker compose logs -f

# Stop containers
docker compose down
```

## Troubleshooting

### Common Issues

#### Step 1: Docker Installation Fails
```bash
# Check if Docker repository was added correctly
ls /etc/apt/sources.list.d/ | grep docker

# Try running the setup script again
chmod +x ./scripts/setup-docker.sh
./scripts/setup-docker.sh

# Or try manual installation
sudo apt-get update
sudo apt-get install docker.io docker-compose

# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in, or run:
newgrp docker

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker
```

#### Step 2: Docker Permission Issues
```bash
# Check if user is in docker group
groups $USER

# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in, or run:
newgrp docker

# Test Docker access
docker run hello-world
```

#### Step 3: SGX Hardware Not Detected
```bash
# Check CPU support
sudo cat /proc/cpuinfo | grep -i sgx

# Check device files
ls -la /dev/sgx/

# If no device files, SGX may not be enabled in BIOS
echo "SGX hardware not available - check BIOS settings"
```

#### Step 4: Enclave Key Generation Fails
```bash
# Check OpenSSL version
openssl version

# Generate key manually
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -pkeyopt rsa_keygen_pubexp:3 -out secrets/enclave-signing-key.pem

# Set proper permissions
chmod 600 secrets/enclave-signing-key.pem
```

#### Step 5: Docker Build Fails
```bash
# Check Docker daemon
docker info

# Clean Docker cache
docker system prune -a

# Rebuild without cache
make docker-build --no-cache

# Check available disk space
df -h
```

#### Step 6: Container Won't Start
```bash
# Check container logs
make docker-logs

# Check SGX device mapping
docker compose exec aleo-oracle-notarization-backend ls -la /dev/sgx/

# Verify enclave key is mounted
docker compose exec aleo-oracle-notarization-backend ls -la /secrets/

# Check Docker daemon logs
sudo journalctl -u docker.service
```

#### Step 7: Application Errors
```bash
# Check application logs
make docker-logs

# Restart container
make docker-stop
make docker-run

# Check resource usage
docker stats

# Check container health
docker compose ps
```

#### Step 8: Monitoring Setup Issues
```bash
# Check if monitoring services are running
sudo systemctl status prometheus
sudo systemctl status alertmanager

# Check monitoring ports
netstat -tlnp | grep -E "(9090|9093)"

# View monitoring logs
sudo journalctl -u prometheus -f
sudo journalctl -u alertmanager -f

# Test monitoring endpoints
curl http://localhost:9090/api/v1/status/targets
curl http://localhost:9093/api/v1/status
```

## Reproducible Builds

### ✅ Docker Deployment Advantages

**Docker deployment guarantees reproducible builds** because:

- **Pinned Base Images** - Uses specific Gramine version (`gramineproject/gramine:stable-jammy`)
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

### For Maximum Reproducibility
To ensure **100% reproducible builds**:

1. **Use identical Docker versions** across environments
2. **Pin all base images** to specific versions
3. **Use deterministic build process** (Make targets)
4. **Verify enclave measurements** match across deployments

## Next Steps

- Read the [Architecture Guide](architecture.md) for technical details
- Check the [Makefile Guide](makefile-guide.md) for additional commands
- Explore the [API Documentation](api-documentation.md) for integration details
- For native deployment, see the [Native Setup Guide](native-setup-guide.md) 