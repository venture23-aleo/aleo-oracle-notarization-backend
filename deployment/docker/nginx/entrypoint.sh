#!/usr/bin/env bash
set -euo pipefail

echo ">> Nginx mTLS proxy starting..."

: "${UPSTREAM_HOST:=aleo-oracle-notarization-backend}"
: "${UPSTREAM_PORT:=8000}"
: "${NGINX_LISTEN_PORT:=8443}"
# Rate/connection/timeouts defaults (applied here so template uses simple ${VAR})
: "${RATE_LIMIT_ZONE_SIZE:=10m}"
: "${RATE_LIMIT_PER_SECOND:=10}"
: "${RATE_LIMIT_BURST:=20}"
: "${RATE_LIMIT_BURST_API:=30}"
: "${CONN_LIMIT_ZONE_SIZE:=10m}"
: "${CONN_LIMIT_PER_KEY:=20}"
: "${CLIENT_MAX_BODY_SIZE:=2m}"
: "${PROXY_CONNECT_TIMEOUT:=5s}"
: "${PROXY_SEND_TIMEOUT:=60s}"
: "${PROXY_READ_TIMEOUT:=60s}"

echo ">> Using upstream http://${UPSTREAM_HOST}:${UPSTREAM_PORT} (listen ${NGINX_LISTEN_PORT})"

TEMPLATE=/etc/nginx/templates/default.conf.template
TARGET=/etc/nginx/conf.d/default.conf

CERT_DIR=/etc/nginx/certs
SERVER_CRT=$CERT_DIR/server.crt
SERVER_KEY=$CERT_DIR/server.key
CA_CRT=$CERT_DIR/ca.crt
REVOKED_DIR=/etc/nginx/revoked

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

# Ensure revoked directory is present (read-only mount expected from compose)
if [ ! -d "$REVOKED_DIR" ]; then
  echo "[WARN] Revocation directory $REVOKED_DIR not found. Revocation checks will accept all clients."
fi

echo ">> Rendering Nginx configuration from template..."
# Include new tunables for rate limiting and timeouts when substituting
envsubst '${UPSTREAM_HOST} ${UPSTREAM_PORT} ${NGINX_LISTEN_PORT} ${RATE_LIMIT_ZONE_SIZE} ${RATE_LIMIT_PER_SECOND} ${RATE_LIMIT_BURST} ${RATE_LIMIT_BURST_API} ${CONN_LIMIT_ZONE_SIZE} ${CONN_LIMIT_PER_KEY} ${CLIENT_MAX_BODY_SIZE} ${PROXY_CONNECT_TIMEOUT} ${PROXY_SEND_TIMEOUT} ${PROXY_READ_TIMEOUT}' < "$TEMPLATE" > "$TARGET"

echo ">> Final rendered config:"
grep -E 'listen|proxy_pass|ssl_' "$TARGET" || true

echo ">> Starting Nginx..."
exec nginx -g 'daemon off;'