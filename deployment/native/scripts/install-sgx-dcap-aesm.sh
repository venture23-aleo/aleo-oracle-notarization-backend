#!/bin/bash

set -euo pipefail

NATIVE_DEPLOYMENT_DIR=${NATIVE_DEPLOYMENT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}

readonly QCNL_CONFIG_FILE=${NATIVE_QCNL_CONFIG_FILE:-${NATIVE_DEPLOYMENT_DIR}/inputs/sgx_default_qcnl.conf}

# Add Intel SGX repository
echo "[+] Adding Intel SGX repository..."
sudo curl -fsSLo /etc/apt/keyrings/intel-sgx-deb.asc https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key
echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/intel-sgx-deb.asc] https://download.01.org/intel-sgx/sgx_repo/ubuntu $(lsb_release -sc) main" \
| sudo tee /etc/apt/sources.list.d/intel-sgx.list

echo "[+] Updating package list..."
sudo apt-get update

echo "[+] Installing Intel SGX DCAP libraries..."
sudo apt-get install libsgx-dcap-default-qpl  libsgx-dcap-ql 

echo "[+] Copying default QCNL configuration file..."
sudo cp $QCNL_CONFIG_FILE /etc/sgx_default_qcnl.conf

echo "[+] Enabling and starting AESMD service..."
sudo systemctl enable aesmd
sudo systemctl start 

echo "[✓] Intel SGX DCAP libraries installation complete."