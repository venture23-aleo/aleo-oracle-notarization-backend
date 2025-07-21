# Native Setup Guide (Non-Docker)

This guide covers how to run the Aleo Oracle Notarization Backend natively on your system using Gramine directly, without Docker containers.

## ⚠️ Important Notice: Reproducibility Limitations

**Native deployment does NOT guarantee reproducible builds** because different Ubuntu releases and system configurations will have different Gramine versions. This means:

- **Enclave measurements (MRENCLAVE) may differ** between environments
- **Build reproducibility is not guaranteed** across different systems
- **Version consistency requires manual management**

**For guaranteed reproducible builds**, use the [Docker Setup Guide](setup-guide.md) instead.

### Deployment Method Comparison

| Aspect | Native Deployment | Docker Deployment |
|--------|------------------|-------------------|
| **Reproducibility** | ❌ Not guaranteed | ✅ Guaranteed |
| **Setup Complexity** | ⚠️ Moderate | ✅ Simple |
| **System Dependencies** | ❌ High | ✅ Minimal |
| **Version Control** | ❌ Manual | ✅ Automated |
| **Enclave Consistency** | ❌ May vary | ✅ Consistent |
| **Development Speed** | ✅ Fast | ⚠️ Slower |

## Prerequisites

### System Requirements
- **Linux distribution** (Ubuntu 20.04+ recommended)
- **Intel SGX hardware** (with DCAP support and drivers installed)
- **Access to SGX device files** (`/dev/sgx/enclave`, `/dev/sgx/provision`)
- **Go 1.24.4** installed
- **Gramine** installed and configured

### Required Tools
- `go` - Go compiler and tools
- `gramine-manifest` - Gramine manifest generator
- `gramine-sgx-sign` - Gramine SGX manifest signer
- `gramine-sgx` - Gramine SGX runtime
- `gramine-direct` - Gramine direct runtime (for non-SGX systems)

## 1. Install Gramine and Intel SGX related packages

### Ubuntu/Debian
```bash
# Add Gramine repository
sudo apt-get install -y wget
wget -qO - https://packages.gramineproject.io/gramine-keyring.gpg | sudo tee /etc/apt/trusted.gpg.d/gramine-keyring.gpg > /dev/null
echo "deb [arch=amd64] https://packages.gramineproject.io/ $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/gramine.list

# Add Intel SGX repository
sudo curl -fsSLo /etc/apt/keyrings/intel-sgx-deb.asc https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key
echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/intel-sgx-deb.asc] https://download.01.org/intel-sgx/sgx_repo/ubuntu $(lsb_release -sc) main" \
| sudo tee /etc/apt/sources.list.d/intel-sgx.list

# Install Gramine
sudo apt-get update
sudo apt-get install -y gramine libsgx-dcap-default-qpl  libsgx-dcap-ql 

sudo cp ./native/inputs/sgx_default_qcnl.conf /etc/sgx_default_qcnl.conf

# Enable and start AESMD
sudo systemctl enable aesmd
sudo systemctl start aesmd

# Verify installation
gramine-sgx --version
```

Please refer to [setup-gramine-sgx.sh](../native/setup-gramine-sgx.sh) for more details.

### Other Distributions
Refer to the [Gramine installation guide](https://gramine.readthedocs.io/en/stable/installation.html) for your specific distribution.

### Verify Installation
After installation, verify that all required components are available:

```bash
# Check Gramine version
gramine-sgx --version

# Verify all required tools are installed
which gramine-manifest
which gramine-sgx-sign
which gramine-sgx
```

## 2. Generate Enclave Signing Key

Generate the RSA 3072-bit private key required for SGX enclave signing:

```bash
# Using Gramine's key generation tool
make gen-key

# OR using OpenSSL (alternative)
make gen-key-openssl
```

This creates `secrets/enclave-key.pem` which will be used to sign the Gramine manifest.

## 3. Build and Run Natively

### Quick Start
```bash
# Build and run in one command
./native/run-native.sh
```

### Step-by-Step Process

#### 3.1 Build the Go Application
```bash
# Build the Go binary
go build -o ./native/outputs/aleo-oracle-notarization-backend ./cmd/server/main.go
```

#### 3.2 Generate Gramine Manifest
```bash
# Generate manifest from template
gramine-manifest \
  ./native/inputs/aleo-oracle-notarization-backend.manifest.template \
  ./native/outputs/aleo-oracle-notarization-backend.manifest
```

#### 3.3 Sign the Manifest
```bash
# Sign the manifest for SGX
gramine-sgx-sign \
  --manifest ./native/outputs/aleo-oracle-notarization-backend.manifest \
  --output ./native/outputs/aleo-oracle-notarization-backend.manifest.sgx
  --key ./secrets/enclave-key.pem
```

#### 3.4 Run the Application
```bash
# Check SGX support and run appropriately
if [ -e /dev/sgx/enclave ] || [ -e /dev/isgx ]; then
    echo "SGX supported. Running with gramine-sgx..."
    gramine-sgx ./native/outputs/aleo-oracle-notarization-backend
else
    echo "SGX not supported. Running with gramine-direct..."
    gramine-direct ./native/outputs/aleo-oracle-notarization-backend
fi
```

## 4. Configuration

### Environment Variables
You can configure the application using environment variables:

```bash
export LOG_LEVEL=debug
export BUILD_CACHE=true
export CLEAN_BUILD=false
export VERBOSE=false
```

### Configuration Files
The application uses configuration files located in `internal/configs/`. Key files:

- `config.json` - Main application configuration

## 5. Development Workflow

### Using Make Commands (Recommended)
The Makefile provides convenient commands for native development:

```bash
# Build and run natively with Gramine
make native-run

# Build only (without running)
make native-build

# Clean native build artifacts
make native-clean
```

### Using the Script Directly
You can also use the `native/run-native.sh` script directly:

```bash
# Basic run
./native/run-native.sh

# Force clean build
CLEAN_BUILD=true ./native/run-native.sh

# Verbose output
VERBOSE=true ./native/run-native.sh

# Disable build cache
BUILD_CACHE=false ./native/run-native.sh
```

## 6. Troubleshooting

### Common Issues

#### SGX Device Not Found
```bash
# Check if SGX device files exist
ls -la /dev/sgx/

#### Gramine Not Found or Wrong Version
```bash
# Check Gramine installation and version
gramine-sgx --version
which gramine-manifest
which gramine-sgx-sign
which gramine-sgx

#### Manifest Generation Errors
```bash
# Check template syntax
cat ./native/inputs/aleo-oracle-notarization-backend.manifest.template

# Verify template exists
ls -la ./native/inputs/*.template
```

#### Build Errors
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

### Debug Mode
Enable verbose logging for debugging:

```bash
# Run with debug output
VERBOSE=true LOG_LEVEL=debug ./native/run-native.sh
```

## 7. Reproducibility and Version Consistency

### ⚠️ Important: Native Deployment Limitations

**Native deployment does NOT guarantee reproducible builds** because:

- **Different Ubuntu releases** (jammy, focal, etc.) have different Gramine versions
- **Package repositories** may contain different Gramine versions
- **System updates** can change Gramine versions over time
- **Manual installations** may use different versions

### Gramine Version Dependency
```bash
# Check current Gramine version
gramine-sgx --version
```

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