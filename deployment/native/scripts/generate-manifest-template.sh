#!/usr/bin/env bash

set -euo pipefail

APP=${APP:-aleo-oracle-notarization-backend}

NATIVE_DEPLOYMENT_DIR=${NATIVE_DEPLOYMENT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}

INPUTS_DIR=${NATIVE_DEPLOYMENT_INPUTS_DIR:-${NATIVE_DEPLOYMENT_DIR}/inputs}

ENCLAVE_ARTIFACTS_DIR=${NATIVE_DEPLOYMENT_ENCLAVE_ARTIFACTS_DIR:-${NATIVE_DEPLOYMENT_DIR}/enclave_artifacts}

MANIFEST_TEMPLATE=${NATIVE_DEPLOYMENT_MANIFEST_TEMPLATE:-${INPUTS_DIR}/${APP}.manifest.template}

LD_LIBRARY_PATH=${LD_LIBRARY_PATH:-/lib/x86_64-linux-gnu/}

mkdir -p "$(dirname "$MANIFEST_TEMPLATE")"

echo ">> Generating native manifest template at $MANIFEST_TEMPLATE..."

# Generate the .manifest.template
cat > $MANIFEST_TEMPLATE <<EOF

[loader.env]

[fs.root]
type = "chroot"
uri = "file:/"

[fs]
mounts = [
  { uri = "file:{{ gramine.runtimedir() }}", path = "/lib" },
  { uri = "file:${LD_LIBRARY_PATH}", path = "${LD_LIBRARY_PATH}" },
  { uri = "file:${ENCLAVE_ARTIFACTS_DIR}/${APP}", path = "/${APP}" },
  { uri = "file:/etc/ssl/", path = "/etc/ssl/" },
  { uri = "file:/usr/lib/ssl/", path = "/usr/lib/ssl/" },
  { uri = "file:/etc/resolv.conf", path = "/etc/resolv.conf" },
  { uri = "file:/etc/hosts", path = "/etc/hosts" },
  { uri = "file:/etc/sgx_default_qcnl.conf", path = "/etc/sgx_default_qcnl.conf" }
]

[sgx]
debug = false
edmm_enable = {{ 'true' if env.get('EDMM', '0') == '1' else 'false' }}
trusted_files= [
  "file:{{ gramine.runtimedir() }}/",
  "file:$LD_LIBRARY_PATH",
  "file:${ENCLAVE_ARTIFACTS_DIR}/${APP}",
  "file:/etc/sgx_default_qcnl.conf",
  "file:/etc/resolv.conf",
  "file:/etc/hosts",
  "file:/etc/ssl/",
  "file:/usr/lib/ssl/"
]
isvprodid = 1
isvsvn = 1
max_threads = 16
enclave_size = "2G"
remote_attestation = "dcap"

[libos]
entrypoint = "/${APP}"
EOF

echo "✅ Gramine manifest template generated as $MANIFEST_TEMPLATE" 