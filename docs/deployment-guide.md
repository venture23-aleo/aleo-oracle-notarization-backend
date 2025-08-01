# 🚀 Deployment Guide

This guide covers how to deploy the SGX Go application using either **Native** or **Docker** deployment methods.

## 📋 Overview

### Quick Comparison

| Aspect | Make Targets | Step-by-Step Manual |
|--------|-------------|-------------------|
| **Speed** | ⚡ Fast (one command) | 🐌 Slower (multiple steps) |
| **Control** | ⚠️ Limited | ✅ Full control |
| **Learning** | ❌ Less educational | ✅ Better understanding |
| **Debugging** | ⚠️ Harder to isolate issues | ✅ Easy to debug steps |
| **Automation** | ✅ Perfect for CI/CD | ❌ Manual process |
| **Recommended for** | First-time setup, development | Learning, troubleshooting |

---

## ⚠️ Important Notice: Reproducible Builds

**Docker deployment guarantees reproducible builds** because all components are containerized and versioned. This means:

- **Enclave measurements (MRENCLAVE) are consistent** across environments
- **Build reproducibility is guaranteed** across different systems
- **Version consistency is automated** through Docker images


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

- **🧪 Testing & Development**: Use **Native** deployment for easier debugging and direct system access
- **🏭 Production & CI/CD**: Use **Docker** deployment for consistent, reproducible environments

## 📦 Deployment Steps

### Docker Deployment
1. Install Docker related components
2. Generate enclave signing key
3. Generate the manifest template
4. Build the Docker image
5. Run the Docker container with SGX support
6. Setup Prometheus and Alertmanager (optional). Please refer to the [Prometheus and Alertmanager Setup Guide](monitoring-setup-guide.md) for more details.

#### Please refer to the [Docker Deployment Guide](docker-deployment-guide.md) for more details.**

### Native Deployment
1. Install SGX DCAP AESM
2. Install Gramine
3. Install Go
4. Generate enclave signing key
5. Generate the manifest template
5. Build and run the native application
6. Setup Prometheus and Alertmanager (optional). Please refer to the [Prometheus and Alertmanager Setup Guide](monitoring-setup-guide.md) for more details.

#### Please refer to the [Native Deployment Guide](native-deployment-guide.md) for more details.**