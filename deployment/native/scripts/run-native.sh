#!/bin/bash

# Enhanced SGX Enclave Runner Script
# Optimized for performance, security, and maintainability

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Configuration
readonly APP_NAME=${APP:-aleo-oracle-notarization-backend}


# Use environment variables with fallback to path calculation
readonly SCRIPT_DIR=${NATIVE_DEPLOYMENT_SCRIPTS_DIR:-"$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"}
readonly NATIVE_DEPLOYMENT_DIR=$(cd "$SCRIPT_DIR/.." && pwd)
readonly DEPLOYMENT_ROOT=$(cd "$NATIVE_DEPLOYMENT_DIR/.." && pwd)
readonly PROJECT_ROOT=$(cd "$DEPLOYMENT_ROOT/.." && pwd)
readonly ENCLAVE_ARTIFACTS_DIR="$NATIVE_DEPLOYMENT_DIR/enclave_artifacts"
readonly INPUT_DIR="$NATIVE_DEPLOYMENT_DIR/inputs"
readonly DEPLOYMENT_DIR=$(cd "$NATIVE_DEPLOYMENT_DIR/.." && pwd)
readonly ENTRYPOINT="$ENCLAVE_ARTIFACTS_DIR/$APP_NAME"
readonly SGX_KEY_FILE="$DEPLOYMENT_ROOT/secrets/enclave-signing-key.pem"


# Environment variables with defaults
readonly LOG_LEVEL="${LOG_LEVEL:-debug}"
readonly BUILD_CACHE="${BUILD_CACHE:-false}"
readonly CLEAN_BUILD="${CLEAN_BUILD:-true}"
readonly VERBOSE="${VERBOSE:-false}"

# Colors for logging
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] INFO:${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARN:${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"
}

log_debug() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')] DEBUG:${NC} $1"
    fi
}

# Error handling
cleanup() {
    log_debug "Cleaning up temporary files..."
    # Add cleanup logic here if needed
}

trap cleanup EXIT

# Validation functions
validate_environment() {
    log_info "Validating environment..."
    
    # Check if we're in the right directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        log_error "go.mod not found. Please run from project root."
        exit 1
    fi
    
    # Check required tools
    local missing_tools=()
    for tool in go gramine-manifest gramine-sgx-sign gramine-sgx gramine-direct gramine-sgx-gen-private-key; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi
    
    # Check SGX support
    if [[ -e /dev/sgx_enclave ]] || [[ -e /dev/isgx ]]; then
        log_info "SGX hardware detected"
        readonly SGX_MODE="sgx"
    else
        log_warn "SGX hardware not detected, will use gramine-direct"
        readonly SGX_MODE="direct"
    fi
}

validate_files() {
    log_info "Validating required files..."
    
    local template_manifest="$INPUT_DIR/$APP_NAME.manifest.template"
    if [[ ! -f "$template_manifest" ]]; then
        log_error "Manifest template not found: $template_manifest"
        exit 1
    fi
    
    # Ensure enclave artifacts directory exists
    mkdir -p "$ENCLAVE_ARTIFACTS_DIR"
}

# Build optimization
should_rebuild() {
    if [[ "$CLEAN_BUILD" == "true" ]]; then
        return 0  # Always rebuild
    fi
    
    if [[ "$BUILD_CACHE" == "false" ]]; then
        return 0  # Always rebuild
    fi
    
    return 1  # No rebuild needed
}

build_application() {
    log_info "Building application..."
    
    if should_rebuild; then
        log_debug "Building $APP_NAME..."
        
        # Use Go build with optimizations
        local build_args=(
            -o "$ENTRYPOINT"
            -ldflags="-s -w"  # Strip debug info for smaller binary
        )
        
        if [[ "$VERBOSE" == "true" ]]; then
            build_args+=(-v)
        fi
        
        if ! go build "${build_args[@]}" "$PROJECT_ROOT/cmd/server/main.go"; then
            log_error "Build failed"
            exit 1
        fi
        
        log_info "Build completed successfully"
    else
        log_info "Using cached build"
    fi
    
    # Verify binary exists and is executable
    if [[ ! -x "$ENTRYPOINT" ]]; then
        log_error "Built binary is not executable: $ENTRYPOINT"
        exit 1
    fi
}

# Manifest generation
generate_manifest() {
    log_info "Generating Gramine manifest..."
    
    local template_manifest="$INPUT_DIR/$APP_NAME.manifest.template"
    local actual_manifest="$ENCLAVE_ARTIFACTS_DIR/$APP_NAME.manifest"
    local signed_manifest="$ENCLAVE_ARTIFACTS_DIR/$APP_NAME.manifest.sgx"
    
    # Clean up old manifests
    log_debug "Removing old manifest files..."
    rm -f "$actual_manifest" "$signed_manifest"
    
    # Generate manifest from template
    log_debug "Generating manifest from template..."
    if ! gramine-manifest "$template_manifest" "$actual_manifest"; then
        log_error "Failed to generate manifest from template"
        exit 1
    fi
    
    # Verify manifest was created
    if [[ ! -f "$actual_manifest" ]]; then
        log_error "Generated manifest not found: $actual_manifest"
        exit 1
    fi
    
    log_info "Manifest generated successfully"
}

# Sign manifest
sign_manifest() {
    log_info "Signing SGX manifest..."
    
    local actual_manifest="$ENCLAVE_ARTIFACTS_DIR/$APP_NAME.manifest"
    local signed_manifest="$ENCLAVE_ARTIFACTS_DIR/$APP_NAME.manifest.sgx"
    
    if ! gramine-sgx-sign --manifest "$actual_manifest" --output "$signed_manifest" --key "$SGX_KEY_FILE"; then
        log_error "Failed to sign SGX manifest"
        exit 1
    fi
    
    if [[ ! -f "$signed_manifest" ]]; then
        log_error "Signed SGX manifest not found: $signed_manifest"
        exit 1
    fi
    
    log_info "Manifest signed successfully"
}

# Run application
run_application() {
    log_info "Starting application..."
    
    if [[ "$SGX_MODE" == "sgx" ]]; then
        log_info "Running with gramine-sgx..."
        exec gramine-sgx "$ENTRYPOINT"
    else
        log_info "Running with gramine-direct..."
        exec gramine-direct "$ENTRYPOINT"
    fi
}

# Main execution
main() {
    log_info "Starting SGX Enclave Runner"
    log_info "Configuration: LOG_LEVEL=$LOG_LEVEL, BUILD_CACHE=$BUILD_CACHE, CLEAN_BUILD=$CLEAN_BUILD, VERBOSE=$VERBOSE"
    
    validate_environment
    validate_files
    build_application
    generate_manifest
    sign_manifest
    run_application
}

# Run main function
main "$@"