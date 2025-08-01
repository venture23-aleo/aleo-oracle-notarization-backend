# Native Deployment Guide (Non-Docker)

This guide covers how to run the Aleo Oracle Notarization Backend natively on your system using Gramine directly, without Docker containers.

## 🚀 Quick Start Options

You have **two approaches** to set up the native environment:

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

---

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

## Prerequisites

### System Requirements
- **Linux distribution** (Ubuntu 20.04+ recommended)
- **Intel SGX hardware** (with DCAP support and drivers installed)
- **Access to SGX device files** (`/dev/sgx/enclave`, `/dev/sgx/provision`)
- **Gramine** installed and configured
- **Go 1.24.4** installed

### Required Tools
- `go` - Go compiler and tools
- `gramine-manifest` - Gramine manifest generator
- `gramine-sgx-sign` - Gramine SGX manifest signer
- `gramine-sgx` - Gramine SGX runtime
- `gramine-direct` - Gramine direct runtime (for non-SGX systems)

## Step-by-Step Installation Process

### Step 1: Check Linux Distribution

First, verify your Linux distribution is compatible:

```bash
# Check Linux distribution and version
lsb_release -a

# Verify it's Ubuntu 20.04 or later
echo "Distribution check completed"
```

**Expected Output:**
```
Distributor ID: Ubuntu
Description:    Ubuntu 22.04.5 LTS
Release:        22.04
Codename:       jammy
```

### Step 2: Check SGX Hardware Support

Verify that your system has SGX hardware support:

```bash
# Check CPU SGX support
sudo cat /proc/cpuinfo | grep -i sgx

# Check if SGX device files exist
ls -la /dev/sgx_enclave 2>/dev/null || echo "SGX device files not found"

# Check AESM service status (if already installed)
sudo systemctl status aesmd --no-pager -l 2>/dev/null || echo "AESM service not found (will be installed in Step 3)"
```

**Expected Output:**
```
# CPU check should show: sgx sgx_lc (in CPU flags)
# Device files should show: crw-rw-rw- 1 root sgx 10, 125 /dev/sgx_enclave
```

### Step 3: Install Intel SGX and DCAP Drivers

Install Intel SGX drivers and DCAP libraries:

```bash
# Run the Intel SGX setup script
make native-install-intel-sgx

# Verify installation
sudo systemctl status aesmd --no-pager -l
```

**Expected Output:**
```
● aesmd.service - Intel(R) Architectural Enclave Service Manager
     Loaded: loaded (/lib/systemd/system/aesmd.service; enabled; vendor preset: enabled)
     Active: active (running)
```

### Step 4: Install Gramine Framework

Install and configure Gramine:

```bash
# Run the Gramine setup script
chmod +x ./native/setup-gramine.sh
./native/setup-gramine.sh

# Verify Gramine installation
gramine-sgx --version

# Check all required Gramine tools
for tool in gramine-manifest gramine-sgx-sign gramine-sgx gramine-direct; do
    if command -v "$tool" &> /dev/null; then
        echo "✅ $tool found: $(which $tool)"
    else
        echo "❌ $tool not found"
    fi
done
```

**Expected Output:**
```
Gramine 1.9 (0d1a4b7607592dab4c8a720c962acee3de6b4ca8)
✅ gramine-manifest found: /usr/bin/gramine-manifest
✅ gramine-sgx-sign found: /usr/bin/gramine-sgx-sign
✅ gramine-sgx found: /usr/bin/gramine-sgx
✅ gramine-direct found: /usr/bin/gramine-direct
```

### Step 5: Install Go and Check Version

Install Go 1.24.4 and verify the installation:

```bash
# Run the Go setup script
chmod +x ./native/setup-go.sh
./native/setup-go.sh

# Verify installation
go version
```

**Expected Output:**
```
go version go1.24.4 linux/amd64
```

### Step 6: Generate Enclave Signing Key

Generate the RSA 3072-bit private key required for SGX enclave signing:

```bash
make generate-enclave-signing-key
```

### Step 7: Run the Application

Now you can build and run the application with Gramine:

```bash
# Build and run the application
./native/run-native.sh
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

## Configuration

### Environment Variables
You can configure the application using environment variables:

```bash
export LOG_LEVEL=debug
export BUILD_CACHE=true
export CLEAN_BUILD=false
```

### Configuration Files
The application uses configuration files located in `internal/configs/`. Key files:

- `config.json` - Main application configuration

## Troubleshooting

### Common Issues

#### Step 2: SGX Hardware Not Detected
```bash
# Check CPU support
sudo cat /proc/cpuinfo | grep -i sgx

# Check device files
ls -la /dev/sgx/

# If no device files, SGX may not be enabled in BIOS
echo "SGX hardware not available - check BIOS settings"
```

#### Step 3: Intel SGX Installation Fails
```bash
# Check if repositories were added correctly
ls /etc/apt/sources.list.d/ | grep sgx

# Try manual installation
sudo apt-get update
sudo apt-get install libsgx-dcap-default-qpl libsgx-dcap-ql

# Check AESM service
sudo systemctl status aesmd

# Or run the setup script manually
chmod +x ./native/setup-intel-dcap-aesm.sh
./native/setup-intel-dcap-aesm.sh
```

#### Step 4: Gramine Installation Fails
```bash
# Check if repositories were added correctly
ls /etc/apt/sources.list.d/ | grep gramine

# Try manual installation
sudo apt-get update
sudo apt-get install gramine

# Verify installation
gramine-sgx --version
```

#### Step 5: Go Installation Issues
```bash
# Check current Go version
go version

# If wrong version, run the setup script
chmod +x ./native/install-go.sh
./native/install-go.sh

# Or install manually
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### Step 6: Application Build/Run Errors
```bash
# Check Go installation
go version

# Clean and rebuild
make clean
make build

# Check for missing dependencies
go mod tidy
go mod download

```

## Reproducibility and Version Consistency

### ⚠️ Important: Native Deployment Limitations

**Native deployment does NOT guarantee reproducible builds** because:

- **Different Ubuntu releases** (jammy, focal, etc.) have different Gramine versions
- **Package repositories** may contain different Gramine versions
- **System updates** can change Gramine versions over time
- **Manual installations** may use different versions

### For Reproducible Builds
If you need **guaranteed reproducible builds** with consistent enclave measurements (MRENCLAVE):

1. **Use Docker deployment** - See [Docker Setup Guide](setup-guide.md)
2. **Pin specific Gramine versions** across all environments
3. **Use identical Ubuntu releases** and package versions
4. **Avoid system updates** that change Gramine

## Next Steps

- Read the [Architecture Guide](architecture.md) for technical details
- Check the [Makefile Guide](makefile-guide.md) for additional commands
- Explore the [API Documentation](api-documentation.md) for integration details 