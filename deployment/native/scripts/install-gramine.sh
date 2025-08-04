#!/bin/bash

set -euo pipefail

echo "[+] Adding Gramine repository..."
# Add Gramine repository
sudo curl -fsSLo /etc/apt/keyrings/gramine-keyring-$(lsb_release -sc).gpg https://packages.gramineproject.io/gramine-keyring-$(lsb_release -sc).gpg
echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/gramine-keyring-$(lsb_release -sc).gpg] https://packages.gramineproject.io/ $(lsb_release -sc) main" \
| sudo tee /etc/apt/sources.list.d/gramine.list

echo "[+] Updating package list..."
sudo apt-get update

echo "[+] Installing Gramine..."
sudo apt-get install -y gramine

echo "[✓] Gramine installation complete."
echo "[✓] Verifying Gramine version:"
gramine-sgx --version

echo "[✓] Gramine installation verified"

for tool in gramine-manifest gramine-sgx-sign gramine-sgx gramine-direct; do
    if command -v "$tool" &> /dev/null; then
        echo "✅ $tool found: $(which $tool)"
    else
        echo "❌ $tool not found"
    fi
done