#!/usr/bin/env bash
set -euo pipefail

# Unified mTLS certificate management script
# Commands:
#   init                - Generate CA + server cert (if not exist) and optionally a default client
#                         Flags: --with-default-client --server-cn CN --server-sans DNS:one,DNS:two,URI:spiffe://id
#   regen-server        - Regenerate only server certificate (same flags as init for CN/SAN override)
#   generate-client     - Create new client cert
#                         Flags: --cn NAME [--days N] [--org ORG] [--country CC] [--san SAN_ENTRY ...]
#                               SAN_ENTRY examples: DNS:client.example.com URI:spiffe://oracle/client/alice email:alice@example.com
#   renew-client        - Replace existing client cert (same flags as generate-client; SANs replace previous)
#   revoke-client       - Mark client as revoked (--cn NAME)
#   list-clients        - List issued client certs (CN + fingerprint)
#   show-client         - Show certificate details (--cn NAME)
#   dump-ca             - Print CA certificate path
#   help                - Show help
#
# Environment / Paths:
#   mtls/
#     ca.crt ca.key
#     server.crt server.key
#     clients/<CN>/{client.crt,client.key,meta.json}
#     revoked/<CN>.revoked

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MTLS_DIR="${SCRIPT_DIR}/mtls"
CLIENTS_DIR="$MTLS_DIR/clients"
REVOKED_DIR="$MTLS_DIR/revoked"
mkdir -p "$MTLS_DIR" "$CLIENTS_DIR" "$REVOKED_DIR"

CA_KEY="$MTLS_DIR/ca.key"
CA_CERT="$MTLS_DIR/ca.crt"
SERVER_KEY="$MTLS_DIR/server.key"
SERVER_CSR="$MTLS_DIR/server.csr"
SERVER_CERT="$MTLS_DIR/server.crt"
OPENSSL_CNF="$MTLS_DIR/openssl.cnf"

color() { local c=$1; shift; printf "\033[%sm%s\033[0m\n" "$c" "$*"; }
err() { color 31 "[ERROR] $*" >&2; }
info() { color 36 "[INFO]  $*"; }
succ() { color 32 "[OK]    $*"; }

fingerprint() { openssl x509 -in "$1" -noout -fingerprint -sha256 | cut -d'=' -f2 | tr -d ':'; }

ensure_ca() {
  if [[ ! -f $CA_CERT || ! -f $CA_KEY ]]; then
    err "CA not initialized. Run: $0 init"
    exit 1
  fi
}

write_openssl_cnf() {
  local server_cn="$1"; shift
  # Remaining args are SAN entries (already in form DNS:...,URI:...,email:... etc)
  local sans=("$@")
  # If no SANs provided, fall back to defaults
  if [[ ${#sans[@]} -eq 0 ]]; then
    sans=(
      'DNS:localhost'
      'DNS:nginx-mtls-proxy'
      'DNS:aleo-oracle-notarization-backend'
    )
  fi
  local san_csv
  local IFS=,; san_csv="${sans[*]}"
  cat > "$OPENSSL_CNF" <<EOF
[ req ]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = v3_ext
distinguished_name = dn

[ dn ]
C = NP
O = Venture23
CN = ${server_cn}

[ v3_ext ]
subjectAltName = ${san_csv}
subjectKeyIdentifier   = hash
basicConstraints = CA:FALSE
keyUsage = digitalSignature,keyEncipherment
extendedKeyUsage = serverAuth,clientAuth
EOF
}

init() {
  local with_default_client=false
  local server_cn="localhost"
  local server_sans=()
  while [[ $# -gt 0 ]]; do
    case $1 in
      --with-default-client) with_default_client=true; shift ;;
      --server-cn) server_cn=$2; shift 2 ;;
      --server-sans)
        IFS=',' read -r -a server_sans <<< "$2"; shift 2 ;;
      *) err "Unknown flag: $1"; exit 1 ;;
    esac
  done
  # Drop any accidental empty elements (e.g. trailing comma)
  if (( ${#server_sans[@]} > 0 )); then
    local cleaned=()
    for s in "${server_sans[@]}"; do [[ -n $s ]] && cleaned+=("$s"); done
    server_sans=("${cleaned[@]}")
  fi
  if [[ -f $CA_CERT ]]; then
    info "CA already exists, skipping CA generation"
  else
    info "Generating Root CA"
    openssl genrsa -out "$CA_KEY" 4096 >/dev/null 2>&1
    openssl req -x509 -new -nodes -key "$CA_KEY" -sha256 -days 365 -subj "/C=NP/O=Venture23/CN=Venture23-Root-CA" -out "$CA_CERT" >/dev/null 2>&1
  fi
  # Safe expansion even if array empty: the +... pattern expands only when set & non-empty
  if (( ${#server_sans[@]} > 0 )); then
    write_openssl_cnf "$server_cn" "${server_sans[@]}"
  else
    write_openssl_cnf "$server_cn"
  fi
  info "Generating / refreshing server certificate (CN=${server_cn})"
  openssl genrsa -out "$SERVER_KEY" 2048 >/dev/null 2>&1
  openssl req -new -key "$SERVER_KEY" -out "$SERVER_CSR" -config "$OPENSSL_CNF" >/dev/null 2>&1
  openssl x509 -req -in "$SERVER_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial -out "$SERVER_CERT" -days 180 -sha256 -extensions v3_ext -extfile "$OPENSSL_CNF" >/dev/null 2>&1
  rm -f "$SERVER_CSR"
  chmod 600 "$CA_KEY" "$SERVER_KEY"
  succ "Server cert ready: $SERVER_CERT"
  if $with_default_client; then
    generate_client --cn AleoOracleClient --days 180 || true
  fi
}

regen_server() {
  ensure_ca
  local server_cn="localhost"; local server_sans=()
  while [[ $# -gt 0 ]]; do
    case $1 in
      --server-cn) server_cn=$2; shift 2 ;;
      --server-sans) IFS=',' read -r -a server_sans <<< "$2"; shift 2 ;;
      *) err "Unknown flag: $1"; exit 1 ;;
    esac
  done
  if (( ${#server_sans[@]} > 0 )); then
    write_openssl_cnf "$server_cn" "${server_sans[@]}"
  else
    write_openssl_cnf "$server_cn"
  fi
  info "Regenerating server certificate (CN=${server_cn})"
  openssl genrsa -out "$SERVER_KEY" 2048 >/dev/null 2>&1
  openssl req -new -key "$SERVER_KEY" -out "$SERVER_CSR" -config "$OPENSSL_CNF" >/dev/null 2>&1
  openssl x509 -req -in "$SERVER_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -out "$SERVER_CERT" -days 180 -sha256 -extensions v3_ext -extfile "$OPENSSL_CNF" >/dev/null 2>&1
  rm -f "$SERVER_CSR"
  succ "Server certificate rotated"
}

generate_client() {
  ensure_ca
  local CN="" DAYS=180 ORG="Venture23" COUNTRY="NP" EXTRA_SANS=()
  while [[ $# -gt 0 ]]; do
    case $1 in
      --cn) CN=$2; shift 2 ;;
      --days) DAYS=$2; shift 2 ;;
      --org) ORG=$2; shift 2 ;;
      --country) COUNTRY=$2; shift 2 ;;
      --san) EXTRA_SANS+=("$2"); shift 2 ;;
      *) err "Unknown flag: $1"; exit 1 ;;
    esac
  done
  [[ -n $CN ]] || { err "--cn required"; exit 1; }
  local dir="$CLIENTS_DIR/$CN" key csr cert meta cfg
  dir="$CLIENTS_DIR/$CN"; mkdir -p "$dir"
  key="$dir/client.key"; csr="$dir/client.csr"; cert="$dir/client.crt"; meta="$dir/meta.json"
  if [[ -f $cert ]]; then
    err "Client already exists, use renew-client"
    exit 1
  fi
  info "Generating client cert for $CN (ORG=${ORG} COUNTRY=${COUNTRY})"
  openssl genrsa -out "$key" 2048 >/dev/null 2>&1
  cfg=$(mktemp)
  {
    echo "[req]"; echo "prompt = no"; echo "distinguished_name = dn";
    if [[ ${#EXTRA_SANS[@]} -gt 0 ]]; then echo "req_extensions = v3_req"; fi
    echo "[dn]"; echo "C = ${COUNTRY}"; echo "O = ${ORG}"; echo "CN = ${CN}";
    if [[ ${#EXTRA_SANS[@]} -gt 0 ]]; then
      echo "[v3_req]";
      local IFS=,; echo "subjectAltName = ${EXTRA_SANS[*]}";
      echo "basicConstraints=CA:FALSE";
      echo "keyUsage = digitalSignature, keyEncipherment";
      echo "extendedKeyUsage=clientAuth";
    fi
  } > "$cfg"
  if [[ ${#EXTRA_SANS[@]} -gt 0 ]]; then
    openssl req -new -key "$key" -config "$cfg" -out "$csr" >/dev/null 2>&1
    openssl x509 -req -in "$csr" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial -out "$cert" -days "$DAYS" -sha256 -extensions v3_req -extfile "$cfg" >/dev/null 2>&1
  else
    openssl req -new -key "$key" -subj "/C=${COUNTRY}/O=${ORG}/CN=${CN}" -out "$csr" >/dev/null 2>&1
    openssl x509 -req -in "$csr" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial -out "$cert" -days "$DAYS" -sha256 >/dev/null 2>&1
  fi
  local fp; fp=$(fingerprint "$cert")
  {
    echo -n '{ "cn": '"\"$CN\""', "fingerprint_sha256": '"\"$fp\""', "issued_at_unix": '$(date +%s)', "expires_in_days": '$DAYS', "revoked": false';
    if [[ ${#EXTRA_SANS[@]} -gt 0 ]]; then
      printf ', "sans": ['
      local first=true; for s in "${EXTRA_SANS[@]}"; do
        $first || printf ','; first=false; printf '"%s"' "$s";
      done; printf ']';
    fi; echo ' }';
  } > "$meta"
  chmod 600 "$key"; rm -f "$csr" "$cfg"
  succ "Client $CN created (fingerprint $fp)"
  echo "curl --cacert $CA_CERT --cert $cert --key $key https://localhost:8443/health"
}

renew_client() {
  ensure_ca
  local CN="" DAYS=180 ORG="Venture23" COUNTRY="NP" EXTRA_SANS=()
  while [[ $# -gt 0 ]]; do
    case $1 in
      --cn) CN=$2; shift 2 ;;
      --days) DAYS=$2; shift 2 ;;
      --org) ORG=$2; shift 2 ;;
      --country) COUNTRY=$2; shift 2 ;;
      --san) EXTRA_SANS+=("$2"); shift 2 ;;
      *) err "Unknown flag: $1"; exit 1 ;;
    esac
  done
  [[ -n $CN ]] || { err "--cn required"; exit 1; }
  local dir="$CLIENTS_DIR/$CN" key csr cert meta cfg
  dir="$CLIENTS_DIR/$CN"; [[ -d $dir ]] || { err "Client does not exist"; exit 1; }
  key="$dir/client.key"; csr="$dir/client.csr"; cert="$dir/client.crt"; meta="$dir/meta.json"
  info "Renewing client cert for $CN"
  openssl genrsa -out "$key" 2048 >/dev/null 2>&1
  cfg=$(mktemp)
  if [[ ${#EXTRA_SANS[@]} -gt 0 ]]; then
    {
      echo "[req]"; echo "prompt = no"; echo "distinguished_name = dn"; echo "req_extensions = v3_req";
      echo "[dn]"; echo "C = ${COUNTRY}"; echo "O = ${ORG}"; echo "CN = ${CN}";
      echo "[v3_req]"; local IFS=,; echo "subjectAltName = ${EXTRA_SANS[*]}"; echo "basicConstraints=CA:FALSE"; echo "keyUsage = digitalSignature, keyEncipherment"; echo "extendedKeyUsage=clientAuth";
    } > "$cfg"
    openssl req -new -key "$key" -config "$cfg" -out "$csr" >/dev/null 2>&1
    openssl x509 -req -in "$csr" -CA "$CA_CERT" -CAkey "$CA_KEY" -out "$cert" -days "$DAYS" -sha256 -extensions v3_req -extfile "$cfg" >/dev/null 2>&1
  else
    openssl req -new -key "$key" -subj "/C=${COUNTRY}/O=${ORG}/CN=${CN}" -out "$csr" >/dev/null 2>&1
    openssl x509 -req -in "$csr" -CA "$CA_CERT" -CAkey "$CA_KEY" -out "$cert" -days "$DAYS" -sha256 >/dev/null 2>&1
  fi
  local fp; fp=$(fingerprint "$cert")
  if command -v jq >/dev/null 2>&1 && [[ -f $meta ]]; then
    tmp="$meta.tmp"; jq '.fingerprint_sha256="'$fp'" | .issued_at_unix='$(date +%s)' | .expires_in_days='$DAYS' | .revoked=false' "$meta" > "$tmp" && mv "$tmp" "$meta"
    if [[ ${#EXTRA_SANS[@]} -gt 0 ]]; then jq '.sans='$(printf '%s\n' "${EXTRA_SANS[@]}" | jq -R . | jq -s . ) "$meta" > "$tmp" && mv "$tmp" "$meta"; fi
  else
    {
      echo -n '{ "cn": '"\"$CN\""', "fingerprint_sha256": '"\"$fp\""', "issued_at_unix": '$(date +%s)', "expires_in_days": '$DAYS', "revoked": false';
      if [[ ${#EXTRA_SANS[@]} -gt 0 ]]; then
        printf ', "sans": ['; local first=true; for s in "${EXTRA_SANS[@]}"; do $first || printf ','; first=false; printf '"%s"' "$s"; done; printf ']';
      fi; echo ' }';
    } > "$meta"
  fi
  chmod 600 "$key"; rm -f "$csr" "$cfg"
  succ "Client $CN renewed (fingerprint $fp)"
}

revoke_client() {
  ensure_ca
  local CN=""
  while [[ $# -gt 0 ]]; do
    case $1 in
      --cn) CN=$2; shift 2 ;;
      *) err "Unknown flag: $1"; exit 1 ;;
    esac
  done
  [[ -n $CN ]] || { err "--cn required"; exit 1; }
  local cert="$CLIENTS_DIR/$CN/client.crt"
  [[ -f $cert ]] || { err "Client not found"; exit 1; }
  local fp; fp=$(fingerprint "$cert")
  echo "$fp" > "$REVOKED_DIR/${CN}.revoked"
  succ "Client $CN revoked (fingerprint $fp). Implement CRL/OCSP to enforce."
}

list_clients() {
  ensure_ca
  ls -1 "$CLIENTS_DIR" 2>/dev/null | while read -r d; do
    local cert="$CLIENTS_DIR/$d/client.crt"
    [[ -f $cert ]] || continue
    echo "$d $(fingerprint "$cert")"
  done
}

show_client() {
  ensure_ca
  local CN=""
  while [[ $# -gt 0 ]]; do
    case $1 in
      --cn) CN=$2; shift 2 ;;
      *) err "Unknown flag: $1"; exit 1 ;;
    esac
  done
  [[ -n $CN ]] || { err "--cn required"; exit 1; }
  local cert="$CLIENTS_DIR/$CN/client.crt"
  [[ -f $cert ]] || { err "Client not found"; exit 1; }
  openssl x509 -in "$cert" -noout -text
}

dump_ca() { ensure_ca; echo "$CA_CERT"; }

help() { grep '^#' "$0" | sed 's/^# \{0,1\}//'; }

cmd=${1:-help}; shift || true

case $cmd in
  init) init "$@" ;;
  regen-server) regen_server "$@" ;;
  generate-client) generate_client "$@" ;;
  renew-client) renew_client "$@" ;;
  revoke-client) revoke_client "$@" ;;
  list-clients) list_clients "$@" ;;
  show-client) show_client "$@" ;;
  dump-ca) dump_ca ;;
  help|--help|-h) help ;;
  *) err "Unknown command: $cmd"; help; exit 1 ;;
esac