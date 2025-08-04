#!/bin/bash

set -euo pipefail
# Use environment variables with fallback to path calculation
DEPLOYMENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

SECRETS_DIR=$DEPLOYMENT_DIR/secrets

function generate_enclave_signing_key() {
    mkdir -p "$SECRETS_DIR"

    if [ -f "$SECRETS_DIR/enclave-signing-key.pem" ]; then
        echo ">> Enclave signing key already exists at $SECRETS_DIR/enclave-signing-key.pem" && \
        return 0
    fi

    echo ">> Generating enclave private key at $SECRETS_DIR/enclave-signing-key.pem using OpenSSL ..."
    openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -pkeyopt rsa_keygen_pubexp:3 -out "$SECRETS_DIR/enclave-signing-key.pem"

    echo ">> Setting permissions for $SECRETS_DIR/enclave-signing-key.pem to 600 ..."
    chmod 600 "$SECRETS_DIR/enclave-signing-key.pem"

    echo ">> Enclave signing key generated successfully!"
}

generate_enclave_signing_key