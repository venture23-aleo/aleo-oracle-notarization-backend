#!/usr/bin/env bash

APP="$1"
MANIFEST_TEMPLATE="$2"
LD_LIBRARY_PATH="$3"

mkdir -p "$(dirname "$MANIFEST_TEMPLATE")"

# Generate the .manifest.template
cat > $MANIFEST_TEMPLATE <<EOF

[loader.env]

[fs.root]
type = "chroot"
uri = "file:/"

[fs]
mounts = [
  { uri = "file:{{ gramine.runtimedir() }}", path = "/lib" },
  { uri = "file:${LD_LIBRARY_PATH}", path = "$LD_LIBRARY_PATH" },
  { uri = "file:${APP}", path = "/${APP}" },
  { uri = "file:/etc/ssl/", path = "/etc/ssl/" },
  { uri = "file:/usr/lib/ssl/", path = "/usr/lib/ssl/" },
  { uri = "file:static_resolv.conf", path = "/etc/resolv.conf" },
  { uri = "file:static_hosts", path = "/etc/hosts" },
  { uri = "file:/etc/sgx_default_qcnl.conf", path = "/etc/sgx_default_qcnl.conf" }
]

[sgx]
debug = false
edmm_enable = {{ 'true' if env.get('EDMM', '0') == '1' else 'false' }}
trusted_files= [
  "file:{{ gramine.runtimedir() }}/",
  "file:$LD_LIBRARY_PATH",
  "file:./${APP}",
  "file:/etc/sgx_default_qcnl.conf",
  "file:./static_resolv.conf",
  "file:./static_hosts",
  "file:/etc/ssl/",
  "file:/usr/lib/ssl/"
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