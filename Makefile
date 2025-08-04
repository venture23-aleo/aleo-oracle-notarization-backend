
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

GOCMD := go
export LD_LIBRARY_PATH=/lib/x86_64-linux-gnu/

# ─────────────────────────────────────────────────────────────────────────────
export COMPOSE_BAKE=true
export DEPLOYMENT_DIR := deployment
export DOCKER_DEPLOYMENT_DIR := $(DEPLOYMENT_DIR)/docker
export NATIVE_DEPLOYMENT_DIR := $(DEPLOYMENT_DIR)/native
export SECRETS_DIR := $(DEPLOYMENT_DIR)/secrets

# Default target
.PHONY: all
all:
	@echo ">> Please specify a target. See 'make help' for available targets."	

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

# Check Ubuntu version
check-linux:
	@lsb_release -a

# Check SGX hardware support
check-sgx:
	@echo ">> Checking SGX hardware support..."
	@echo "CPU SGX Support:"
	@sudo cat /proc/cpuinfo | grep -i sgx || echo "SGX not detected in CPU flags"
	@echo "SGX Device Files:"
	@ls -la /dev/sgx_enclave 2>/dev/null || echo "SGX device files not found"
	@echo ">> SGX hardware check completed"

# ─────────────────────────────────────────────────────────────────────────────
# Secrets
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: generate-enclave-signing-key
generate-enclave-signing-key:
	@echo ">> Generating enclave signing key..."
	@chmod +x $(SECRETS_DIR)/generate-enclave-signing-key.sh
	@$(SECRETS_DIR)/generate-enclave-signing-key.sh

# ─────────────────────────────────────────────────────────────────────────────
# Native Deployment
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: native-setup
native-setup:
	@echo ">> Setting up native deployment..."
	@make -C $(NATIVE_DEPLOYMENT_DIR) native-setup

.PHONY: native-run
native-run:
	@echo ">> Running native deployment..."
	@make -C $(NATIVE_DEPLOYMENT_DIR) native-run

# ─────────────────────────────────────────────────────────────────────────────
# Docker Deployment
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: docker-setup
docker-setup:
	@echo ">> Setting up docker deployment..."
	@make -C $(DOCKER_DEPLOYMENT_DIR) docker-setup

.PHONY: docker-build	
docker-build:
	@echo ">> Building docker deployment..."
	@make -C $(DOCKER_DEPLOYMENT_DIR) docker-build

.PHONY: docker-run
docker-run:
	@echo ">> Running docker deployment..."
	@make -C $(DOCKER_DEPLOYMENT_DIR) docker-run

.PHONY: extract-enclave-artifacts
extract-enclave-artifacts:
	@echo ">> Extracting enclave artifacts for docker deployment..."
	@make -C $(DOCKER_DEPLOYMENT_DIR) extract-enclave-artifacts


# ─────────────────────────────────────────────────────────────────────────────
# Monitoring
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: monitoring-setup
monitoring-setup:
	@echo ">> Setting up monitoring tools..."
	@make -C $(MONITORING_DIR) setup

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
	@echo "System Check Targets:"
	@echo "  check-linux     Check Linux distribution compatibility"
	@echo "  check-sgx       Check SGX hardware support"
	@echo
	@echo "Secrets Targets:"
	@echo "  generate-enclave-signing-key    Generate enclave signing key"
	@echo
	@echo "Native Deployment Targets:"
	@echo "  native-setup    Complete native environment setup"
	@echo "  native-start    Run native deployment"
	@echo
	@echo "Docker Deployment Targets:"
	@echo "  docker-setup    Complete Docker environment setup"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-start    Run docker deployment"
	@echo "  extract-enclave-artifacts    Extract enclave artifacts for Docker"
	@echo
	@echo "Monitoring Targets:"
	@echo "  monitoring-setup    Complete monitoring setup"
	@echo
	@echo "  help            Show this help"
