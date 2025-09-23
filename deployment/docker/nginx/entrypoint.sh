#!/usr/bin/env bash
set -euo pipefail

echo ">> Nginx mTLS proxy starting..."

: "${UPSTREAM_HOST:=aleo-oracle-notarization-backend}"
: "${UPSTREAM_PORT:=8000}"
: "${NGINX_LISTEN_PORT:=8443}"

echo ">> Using upstream http://${UPSTREAM_HOST}:${UPSTREAM_PORT} (listen ${NGINX_LISTEN_PORT})"

TEMPLATE=/etc/nginx/templates/default.conf.template
TARGET=/etc/nginx/conf.d/default.conf

CERT_DIR=/etc/nginx/certs
SERVER_CRT=$CERT_DIR/server.crt
SERVER_KEY=$CERT_DIR/server.key
CA_CRT=$CERT_DIR/ca.crt

if [ ! -f "$SERVER_CRT" ] || [ ! -f "$SERVER_KEY" ]; then
  echo "[FATAL] Missing server certificate or key in $CERT_DIR; expected server.crt & server.key" >&2
  echo "[DEBUG] Current contents of $CERT_DIR:" >&2
  ls -l "$CERT_DIR" >&2 || true
  echo "[HINT] Ensure you generated certificates before starting containers:" >&2
  echo "       bash deployment/secrets/generate-mtls-certs.sh init --with-default-client" >&2
  echo "[HINT] Compose relative path mounts server certs from deployment/secrets/mtls. Run this to verify:" >&2
  echo "       ls -l deployment/secrets/mtls/server.crt deployment/secrets/mtls/server.key" >&2
  exit 1
fi

if [ ! -f "$CA_CRT" ]; then
  echo "[FATAL] Missing CA certificate (ca.crt) for client verification." >&2
  ls -l "$CERT_DIR" >&2 || true
  exit 1
fi

echo ">> Rendering Nginx configuration from template..."
envsubst '${UPSTREAM_HOST} ${UPSTREAM_PORT} ${NGINX_LISTEN_PORT}' < "$TEMPLATE" > "$TARGET"

echo ">> Final rendered config:"
grep -E 'listen|proxy_pass|ssl_' "$TARGET" || true

echo ">> Starting Nginx..."
exec nginx -g 'daemon off;'