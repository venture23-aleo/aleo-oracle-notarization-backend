#!/bin/bash

APP="$1"
MANIFEST_TEMPLATE="$2"
LD_LIBRARY_PATH="$3"

mkdir -p "$(dirname "$MANIFEST_TEMPLATE")"

# Generate the .manifest.template
cat > $MANIFEST_TEMPLATE <<EOF

[loader.env]
FORCE_COLOR = "1"
LD_LIBRARY_PATH="$LD_LIBRARY_PATH"
PORT = "$PORT"
WHITELISTED_DOMAINS = $WHITELISTED_DOMAINS

[fs.root]
type = "chroot"
uri = "file:/"

[fs]
mounts = [
  { uri = "file:{{ gramine.runtimedir() }}", path = "/lib" },
  { uri = "file:${LD_LIBRARY_PATH}", path = "$LD_LIBRARY_PATH" },
  { uri = "file:aleo-oracle-notarizer-dev", path = "/${APP}" },
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
  "file:./aleo-oracle-notarizer-dev",
  "file:/etc/sgx_default_qcnl.conf"
]
allowed_files = [
  "file:/etc/resolv.conf",
  "file:/etc/hosts"
]
isvprodid = 1
isvsvn = 1
max_threads = 16
enclave_size = "2G"
remote_attestation = "dcap"

[libos]
entrypoint = "/$APP"
EOF

echo "✅ Gramine manifest template generated as $MANIFEST_TEMPLATE" 