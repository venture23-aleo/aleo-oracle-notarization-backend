#!/bin/bash

# Install Go 1.24.4 for native SGX development
# This script installs Go 1.24.4 from the official binary distribution

set -euo pipefail
readonly GO_VERSION=${GO_VERSION:-1.24.4}
readonly GO_ARCH=${GO_ARCH:-linux-amd64}
readonly GO_INSTALL_DIR=${GO_INSTALL_DIR:-/usr/local}
readonly GO_URL=https://go.dev/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz

echo ">> Installing Go ${GO_VERSION}..."

# Check if Go is already installed with correct version
if command -v go &> /dev/null; then
    CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$CURRENT_VERSION" == "$GO_VERSION" ]]; then
        echo ">> Go ${GO_VERSION} is already installed"
        go version
        exit 0
    else
        echo ">> Found Go ${CURRENT_VERSION}, updating to ${GO_VERSION}..."
    fi
fi

# Download Go binary
echo ">> Downloading Go ${GO_VERSION}..."
wget -q "$GO_URL" -O "/tmp/go${GO_VERSION}.${GO_ARCH}.tar.gz"

# Extract to /usr/local
echo ">> Extracting Go to ${GO_INSTALL_DIR}..."
sudo tar -C "$GO_INSTALL_DIR" -xzf "/tmp/go${GO_VERSION}.${GO_ARCH}.tar.gz"

# Add to PATH if not already there
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    echo ">> Adding Go to PATH in ~/.bashrc..."
    echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
fi

# Export PATH for current session
export PATH=$PATH:/usr/local/go/bin

# Clean up
rm -f "/tmp/go${GO_VERSION}.${GO_ARCH}.tar.gz"

# Verify installation
echo ">> Verifying Go installation..."
go version

echo ">> Go ${GO_VERSION} installation completed successfully!"

source ~/.bashrc

echo ">> Note: Added Go to PATH in ~/.bashrc"