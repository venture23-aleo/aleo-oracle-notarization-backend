#!/bin/bash

set -euo pipefail

APP=${APP:-aleo-oracle-notarization-backend}

DEPLOYMENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

ENCLAVE_ARTIFACTS_DIR=${DEPLOYMENT_DIR}/enclave_artifacts

mkdir -p "$ENCLAVE_ARTIFACTS_DIR"

echo ">> Getting enclave info... at $ENCLAVE_ARTIFACTS_DIR"
echo ">> Creating temporary container..."
container_id=$(docker create ${APP})
echo ">> Container ID: $container_id"
echo ">> Copying enclave artifacts..."
mkdir -p "$ENCLAVE_ARTIFACTS_DIR"

# Copy the Go binary (executable)
docker cp $container_id:/app/${APP} $ENCLAVE_ARTIFACTS_DIR/${APP}
# Copy enclave signature files
docker cp $container_id:/app/${APP}.manifest $ENCLAVE_ARTIFACTS_DIR/${APP}.manifest
docker cp $container_id:/app/${APP}.manifest.sgx $ENCLAVE_ARTIFACTS_DIR/${APP}.manifest.sgx
docker cp $container_id:/app/${APP}.sig  $ENCLAVE_ARTIFACTS_DIR/${APP}.sig
docker cp $container_id:/app/${APP}.metadata.json $ENCLAVE_ARTIFACTS_DIR/${APP}.metadata.json

echo ">> Removing temporary container..."
docker rm $container_id
echo ">> Enclave artifacts extracted successfully! at $ENCLAVE_ARTIFACTS_DIR"
echo ">> Contents:"
ls -la "$ENCLAVE_ARTIFACTS_DIR/"
