# 🚀 Deployment Guide

This guide explains how to deploy the SGX Go application using either the **Native** or **Docker** deployment methods. The project leverages `make` for automation and orchestration. **Before proceeding, please ensure that both `make` is installed and SGX hardware is available.**

Before proceding, please ensure that you have the following prerequisites:
- **Linux distribution** (Ubuntu 20.04+ recommended for host)
- **Intel SGX hardware**
- **Access to SGX device files** (`/dev/sgx_enclave`, `/dev/sgx_provision`)
- **make** - `make` for orchestration


To install `make` on Ubuntu, run the following command:
```bash
sudo apt update && sudo apt install -y make
```

To check if Linux distribution is compatible, run the following command:
```bash
# Check Linux distribution and version
make check-linux
```

To check if SGX hardware is available, run the following command:
```bash
# Check CPU SGX support
make check-sgx
```

### Deployment Method Comparison

| Aspect | Docker Deployment | Native Deployment |
|--------|------------------|-------------------|
| **Reproducibility** | ✅ Guaranteed | ❌ Not guaranteed |
| **Setup Complexity** | ✅ Simple | ⚠️ Moderate |
| **System Dependencies** | ✅ Minimal | ❌ High |
| **Version Control** | ✅ Automated | ❌ Manual |
| **Enclave Consistency** | ✅ Consistent | ❌ May vary |
| **Development Speed** | ⚠️ Slower | ✅ Fast |

### **🎯 Recommendation**

- **🧪 Testing & Development**: Use **Native** deployment for easier debugging and faster development
- **🏭 Production & CI/CD**: Use **Docker** deployment for consistent, reproducible environments and easy deployment

## 📦 Deployment Steps

### Docker Deployment
1. Install Docker and related components (Docker, Docker Compose, BuildKit)
2. Generate enclave signing private key
3. Generate the manifest template
4. Build the Docker image
5. Run the Docker container
6. Setup Prometheus and Alertmanager (optional). Please refer to the [Prometheus and Alertmanager Setup Guide](monitoring-setup-guide.md) for more details.

#### Please refer to the [Docker Deployment Guide](docker-deployment-guide.md) for more details.**

### Native Deployment
1. Install SGX DCAP drivers and start AESM service
2. Install Gramine
3. Install Go (1.24.4)
4. Generate enclave signing key
5. Generate the manifest template
6. Build and run the backend application inside the enclave
7. Setup Prometheus and Alertmanager (optional). Please refer to the [Prometheus and Alertmanager Setup Guide](monitoring-setup-guide.md) for more details.

#### Please refer to the [Native Deployment Guide](native-deployment-guide.md) for more details.**
