
# Makefile for Aleo Oracle Notarizer Development Environment

# Pointing to env. 
# If .env file does not exist, create it
ifeq ($(wildcard .env),)
    $(shell touch .env)
endif

export APP := aleo-oracle-notarization-backend

include .env
export


# ─────────────────────────────────────────────────────────────────────────────
# Variables
# ─────────────────────────────────────────────────────────────────────────────
SHELL := /bin/bash
# export VERSION := $(shell git describe --tags --always --dirty)
export COMMIT := $(shell git rev-parse HEAD)


# LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)"
GOCMD := go
export LD_LIBRARY_PATH=/lib/x86_64-linux-gnu/

# ─────────────────────────────────────────────────────────────────────────────
# Directories
# ─────────────────────────────────────────────────────────────────────────────
export PROJECT_ROOT := $(shell pwd)

export DEPLOYMENT_DIR := $(PROJECT_ROOT)/deployment

export SHARED_DEPLOYMENT_DIR := $(DEPLOYMENT_DIR)/shared
export SHARED_DEPLOYMENT_SCRIPTS_DIR := $(SHARED_DEPLOYMENT_DIR)/scripts
export SHARED_DEPLOYMENT_CONFIGS_DIR := $(SHARED_DEPLOYMENT_DIR)/configs

export DOCKER_DEPLOYMENT_DIR := $(DEPLOYMENT_DIR)/docker
export DOCKER_DEPLOYMENT_SCRIPTS_DIR := $(DOCKER_DEPLOYMENT_DIR)/scripts
export DOCKER_DEPLOYMENT_INPUTS_DIR := $(DOCKER_DEPLOYMENT_DIR)/inputs
export DOCKER_DEPLOYMENT_MANIFEST_TEMPLATE := $(DOCKER_DEPLOYMENT_INPUTS_DIR)/$(APP).manifest.template
export DOCKER_ENCLAVE_ARTIFACTS_DIR := $(DOCKER_DEPLOYMENT_DIR)/enclave_artifacts

export NATIVE_DEPLOYMENT_DIR := $(DEPLOYMENT_DIR)/native
export NATIVE_DEPLOYMENT_SCRIPTS_DIR := $(NATIVE_DEPLOYMENT_DIR)/scripts
export NATIVE_DEPLOYMENT_INPUTS_DIR := $(NATIVE_DEPLOYMENT_DIR)/inputs
export NATIVE_DEPLOYMENT_ENCLAVE_ARTIFACTS_DIR := $(NATIVE_DEPLOYMENT_DIR)/enclave_artifacts
export NATIVE_DEPLOYMENT_MANIFEST_TEMPLATE := $(NATIVE_DEPLOYMENT_INPUTS_DIR)/$(APP).manifest.template

export NATIVE_QCNL_CONFIG_FILE := $(NATIVE_DEPLOYMENT_INPUTS_DIR)/sgx_default_qcnl.conf
export DOCKER_QCNL_CONFIG_FILE := $(DOCKER_DEPLOYMENT_INPUTS_DIR)/sgx_default_qcnl.conf

export SECRETS_DIR := $(DEPLOYMENT_DIR)/secrets
export ENCLAVE_SIGNING_KEY_FILE := $(SECRETS_DIR)/enclave-signing-key.pem


# ─────────────────────────────────────────────────────────────────────────────
# Go version
# ─────────────────────────────────────────────────────────────────────────────
export GO_VERSION := 1.24.4
export GO_ARCH := linux-amd64
export GO_INSTALL_DIR := /usr/local

# ─────────────────────────────────────────────────────────────────────────────
export COMPOSE_BAKE=true

# Default target
.PHONY: all
all: fmt vet lint test build

# ─────────────────────────────────────────────────────────────────────────────
# Test 
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: test
test:
	@echo ">> Running unit tests..."
	$(GOCMD) test -v -race -timeout 30s ./...

# ─────────────────────────────────────────────────────────────────────────────
# Formatting
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: fmt
fmt:
	@echo ">> Formatting code..."
	$(GOCMD) fmt ./...

# ─────────────────────────────────────────────────────────────────────────────
# Vet
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: vet
vet:
	@echo ">> Vetting code..."
	$(GOCMD) vet ./...

# ─────────────────────────────────────────────────────────────────────────────
# Lint
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: lint
lint:
	@echo ">> Static analysis (staticcheck)..."
	@$(GOCMD) install honnef.co/go/tools/cmd/staticcheck@latest
	@$(GOCMD) run honnef.co/go/tools/cmd/staticcheck@latest ./...


# ────────────────────────────────────────────────────────────
# Generate the private key for SGX signing using OpenSSL (exponent 3)
# Usage: make generate-enclave-signing-key
# Requires OpenSSL 1.1.1 or later
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: generate-enclave-signing-key
generate-enclave-signing-key:
	@chmod +x $(SHARED_DEPLOYMENT_SCRIPTS_DIR)/generate-enclave-signing-key.sh
	@$(SHARED_DEPLOYMENT_SCRIPTS_DIR)/generate-enclave-signing-key.sh

# ─────────────────────────────────────────────────────────────────────────────
# Generate manifest template for Docker
# Usage: make generate-docker-manifest-template
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: generate-docker-manifest-template
generate-docker-manifest-template:
	@chmod +x $(DOCKER_DEPLOYMENT_SCRIPTS_DIR)/generate-manifest-template.sh
	@$(DOCKER_DEPLOYMENT_SCRIPTS_DIR)/generate-manifest-template.sh $(APP)

# ─────────────────────────────────────────────────────────────────────────────
# Generate manifest template for Native
# Usage: make generate-native-manifest-template
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: generate-native-manifest-template
generate-native-manifest-template:
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/generate-manifest-template.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/generate-manifest-template.sh $(APP)

# ─────────────────────────────────────────────────────────────────────────────
# Extract enclave artifacts for Docker
# ─────────────────────────────────────────────────────────────────────────────
# Usage: make extract-enclave-artifacts
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: extract-enclave-artifacts
extract-enclave-artifacts:
	@chmod +x $(DOCKER_DEPLOYMENT_SCRIPTS_DIR)/extract-enclave-artifacts.sh
	@$(DOCKER_DEPLOYMENT_SCRIPTS_DIR)/extract-enclave-artifacts.sh $(APP)


# ─────────────────────────────────────────────────────────────────────────────
# Native Installation and Setup
# Usage: make native-setup
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: native-check-linux
native-check-linux:
	@lsb_release -a

.PHONY: native-check-sgx
native-check-sgx:
	@echo ">> Checking SGX hardware support..."
	@echo "CPU SGX Support:"
	@sudo cat /proc/cpuinfo | grep -i sgx || echo "SGX not detected in CPU flags"
	@echo "SGX Device Files:"
	@ls -la /dev/sgx_enclave 2>/dev/null || echo "SGX device files not found"
	@echo ">> SGX hardware check completed"

.PHONY: native-install-sgx-dcap-aesm
native-install-sgx-dcap-aesm:
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-sgx-dcap-aesm.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-sgx-dcap-aesm.sh
	
.PHONY: native-install-gramine
native-install-gramine:
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-gramine.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-gramine.sh

.PHONY: native-install-go
native-install-go:
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-go.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-go.sh $(GO_VERSION)

.PHONY: native-setup
native-setup: native-check-linux native-check-sgx native-install-sgx-dcap-aesm native-install-gramine native-install-go
	@echo ">> Native setup completed successfully!"

# ─────────────────────────────────────────────────────────────────────────────
# Native Build and Run
# Usage: make native-build
# Requires Gramine to be installed
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: native-clean
native-clean:
	@echo ">> Cleaning native build artifacts..."
	@rm -rf $(NATIVE_DEPLOYMENT_ENCLAVE_ARTIFACTS_DIR)/*

.PHONY: native-build
native-build: native-clean
	@echo ">> Building native binary..."
	@$(GOCMD) build -o $(NATIVE_DEPLOYMENT_ENCLAVE_ARTIFACTS_DIR)/$(APP) ./cmd/server

.PHONY: native-run
native-run: native-build generate-native-manifest-template generate-enclave-signing-key
	@echo ">> Running native application with Gramine..."
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/run-native.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/run-native.sh $(APP)

# ─────────────────────────────────────────────────────────────────────────────
# Docker Management (Build, Run, Stop, Logs, Status)
# ─────────────────────────────────────────────────────────────────────────────
# Docker compose flags (can be overridden from command line)
DOCKER_FLAGS ?= -d
DOCKER_COMPOSE_FILE ?= "$(DOCKER_DEPLOYMENT_DIR)/docker-compose.yml"
DOCKER_SERVICES ?= $(APP)

# ─────────────────────────────────────────────────────────────────────────────
# Docker Installation and Setup
# Usage: make docker-setup
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: docker-install
docker-install:
	@chmod +x $(DOCKER_DEPLOYMENT_SCRIPTS_DIR)/install-docker.sh
	@$(DOCKER_DEPLOYMENT_SCRIPTS_DIR)/install-docker.sh

.PHONY: docker-setup 
docker-setup: docker-install generate-enclave-signing-key
	@echo ">> Docker setup completed successfully!"

.PHONY: docker-build
docker-build: generate-docker-manifest-template
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_DEPLOYMENT_INPUTS_DIR="$(DOCKER_DEPLOYMENT_INPUTS_DIR)" ENCLAVE_SIGNING_KEY_FILE="$(ENCLAVE_SIGNING_KEY_FILE)" docker compose -f $(DOCKER_COMPOSE_FILE) build $(APP)
	@docker tag $(APP):latest $(APP):$(COMMIT)

.PHONY: docker-run
docker-run: docker-build
	@echo ">> Running Docker container..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES) $(DOCKER_FLAGS)

.PHONY: docker-run-fg
docker-run-fg: docker-build
	@echo ">> Running Docker container in foreground..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES)

.PHONY: docker-run-rebuild
docker-run-rebuild: docker-build
	@echo ">> Running Docker container in foreground with rebuild..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES) --build --force-recreate

.PHONY: docker-stop
docker-stop:
	@echo ">> Stopping all Docker containers..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) down

.PHONY: docker-logs
docker-logs:
	@echo ">> Showing Docker logs..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) logs -f

.PHONY: docker-status
docker-status:
	@echo ">> Docker container status:"
	@docker compose -f $(DOCKER_COMPOSE_FILE) ps

# ─────────────────────────────────────────────────────────────────────────────
# Prometheus Monitoring And Alertmanager Setup
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: setup-alertmanager
setup-alertmanager:
	@chmod +x $(SHARED_DEPLOYMENT_SCRIPTS_DIR)/setup-alertmanager.sh
	@$(SHARED_DEPLOYMENT_SCRIPTS_DIR)/setup-alertmanager.sh

.PHONY: setup-prometheus
setup-prometheus:
	@chmod +x $(SHARED_DEPLOYMENT_SCRIPTS_DIR)/setup-prometheus.sh
	@$(SHARED_DEPLOYMENT_SCRIPTS_DIR)/setup-prometheus.sh

.PHONY: setup-monitoring
setup-monitoring: setup-alertmanager setup-prometheus
	@echo ">> Monitoring setup complete!"

# ─────────────────────────────────────────────────────────────────────────────
# Help
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make [target] [VARIABLE=value]"
	@echo
	@echo "Development Targets:"
	@echo "  all             Run fmt, vet, lint, test, build"
	@echo "  test            Run tests"
	@echo "  fmt             Format code"
	@echo "  vet             Static vetting"
	@echo "  lint            Static analysis (staticcheck)"
	@echo "  clean           Remove binaries"
	@echo
	@echo "Native Targets:"
	@echo "  native-setup           Complete native environment setup (recommended)"
	@echo "  native-check-linux     Check Linux distribution compatibility"
	@echo "  native-check-sgx       Check SGX hardware support"
	@echo "  native-install-sgx-dcap-aesm Install Intel SGX and DCAP drivers"
	@echo "  native-install-gramine Install Gramine framework"
	@echo "  native-install-go      Install Go 1.24.4"
	@echo "  native-build           Build native binary"
	@echo "  native-run             Build and run native application with Gramine"
	@echo "  native-clean           Clean native build artifacts"
	@echo
	@echo "Docker Targets:"
	@echo "  docker-setup    Complete Docker environment setup (recommended)"
	@echo "  docker-install  Install Docker and Docker Compose"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-run      Build and run Docker container (detached mode)"
	@echo "  docker-run-fg   Build and run Docker container (foreground mode)"
	@echo "  docker-run-rebuild  Build and run Docker container (force rebuild)"
	@echo
	@echo "Docker Variables (can be overridden):"
	@echo "  DOCKER_FLAGS    Docker compose flags (default: -d)"
	@echo "  DOCKER_SERVICES Services to run (default: aleo-oracle-notarization-backend)"
	@echo "  APP             Application name (default: aleo-oracle-notarization-backend)"
	@echo
	@echo "Docker Usage Examples:"
	@echo "  make docker-run                                    # Default detached mode"
	@echo "  make docker-run DOCKER_FLAGS=\"\"                  # Foreground mode"
	@echo "  make docker-run DOCKER_FLAGS=\"--build\"           # With build flag"
	@echo "  make docker-run DOCKER_FLAGS=\"--scale app=2\"     # With scaling"
	@echo "  make docker-run DOCKER_SERVICES=\"app db\"         # Specific services"
	@echo ""
	@echo "Docker Management:"
	@echo "  make docker-stop                                   # Stop all containers"
	@echo "  make docker-logs                                   # Show logs"
	@echo "  make docker-status                                 # Show container status"
	@echo
	@echo "Monitoring Targets:"
	@echo "  setup-monitoring            Setup Alertmanager and Prometheus"
	@echo "  setup-alertmanager          Setup Alertmanager"
	@echo "  setup-prometheus            Setup Prometheus"
	@echo
	@echo "SGX/Enclave Targets:"
	@echo "  generate-docker-manifest-template   Generate manifest template"
	@echo "  ENCLAVE_ARTIFACTS_DIRclave-artifacts    Extract enclave artifacts"
	@echo "  generate-enclave-signing-key Generate enclave signing key (OpenSSL)"
	@echo
	@echo "Utility:"
	@echo "  help            Show this help"
