# Native Deployment Guide (Non-Docker)

This guide covers how to run the Aleo Oracle Notarization Backend natively on your system using Gramine directly, without Docker containers.


## Prerequisites

- **Linux distribution** (Ubuntu 20.04+ recommended)
- **Intel SGX hardware** (with DCAP support and drivers installed)
- **Access to SGX device files** (`/dev/sgx/enclave`, `/dev/sgx/provision`)
- **make** - `make` for orchestration

## 🚀 Quick Start Options

We have **two approaches** to set up the native environment:

### Option 1: Automated Setup (Recommended)
**Use Make targets for quick, automated installation:**
```bash
# Complete setup in one command
make native-setup

# Then run the application
make native-run
```

### Option 2: Step-by-Step Manual Setup
**Follow the detailed step-by-step process below for more control and understanding.**

## Step-by-Step Installation Process

**Note:** Please refer to Makefile inside the `deployment/native` directory for the full list of targets.

### Step 1: Install Intel SGX and DCAP Drivers

Install Intel SGX drivers and DCAP libraries:

```bash
# Run the Intel SGX setup script
make install-sgx-dcap-aesm

# Verify installation
make status-aesmd
```

**Expected Output:**
```
● aesmd.service - Intel(R) Architectural Enclave Service Manager
     Loaded: loaded (/lib/systemd/system/aesmd.service; enabled; vendor preset: enabled)
     Active: active (running)
```

### Step 2: Install Gramine Framework

Install and configure Gramine:

```bash
# Run the Gramine setup script
make install-gramine
```

**Expected Output:**
```
Gramine 1.9 (0d1a4b7607592dab4c8a720c962acee3de6b4ca8)
✅ gramine-manifest found: /usr/bin/gramine-manifest
✅ gramine-sgx-sign found: /usr/bin/gramine-sgx-sign
✅ gramine-sgx found: /usr/bin/gramine-sgx
✅ gramine-direct found: /usr/bin/gramine-direct
```

### Step 3: Install Go and Check Version

Install Go 1.24.4 and verify the installation:

```bash
# Run the Go setup script
make install-go
```

**Expected Output:**
```
go version go1.24.4 linux/amd64
```

### Step 4: Generate Enclave Signing Key

Generate the RSA 3072-bit private key required for SGX enclave signing:

```bash
make generate-enclave-signing-key
```

### Step 5: Generate Manifest Template

Generate the Gramine manifest template:

```bash
make generate-manifest-template
```

### Step 6: Run the Application

Now you can build and run the application with Gramine:

```bash
make native-run
```

**Expected Output:**
```
[2025-07-31 10:32:24] INFO: Starting SGX Enclave Runner
[2025-07-31 10:32:24] INFO: Configuration: LOG_LEVEL=debug, BUILD_CACHE=false, CLEAN_BUILD=true, VERBOSE=false
[2025-07-31 10:32:25] INFO: Validating environment...
[2025-07-31 10:32:25] INFO: SGX hardware detected
[2025-07-31 10:32:25] INFO: Validating required files...
[2025-07-31 10:32:25] INFO: Building application...
[2025-07-31 10:32:25] INFO: Build completed successfully
[2025-07-31 10:32:25] INFO: Generating Gramine manifest...
[2025-07-31 10:32:26] INFO: Manifest generated successfully
[2025-07-31 10:32:26] INFO: Signing SGX manifest...
[2025-07-31 10:32:27] INFO: Manifest signed successfully
[2025-07-31 10:32:27] INFO: Starting application...
[2025-07-31 10:32:27] INFO: Running with gramine-sgx...
Gramine is starting. Parsing TOML manifest file, this may take some time...
```

### Configuration Files
The application uses configuration files located in `internal/configs/`. Key files:

- `config.json` - Main application configuration


### ⚠️ Important: Native Deployment Limitations

**Native deployment does NOT guarantee reproducible builds** because:

- **Different Ubuntu releases** (jammy, focal, etc.) have different Gramine versions
- **Package repositories** may contain different Gramine versions
- **System updates** can change Gramine versions over time
- **Manual installations** may use different versions

## Next Steps

- Read the [Architecture Guide](architecture.md) for technical details
- Explore the [API Documentation](api-documentation.md) for integration details 
- Explore the [Monitoring Setup Guide](monitoring-setup-guide.md) for monitoring and alerting details
- Explore the [Error Codes Reference](error-codes.md) for error codes and troubleshooting