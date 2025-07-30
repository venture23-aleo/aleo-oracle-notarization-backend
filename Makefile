
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
LD_LIBRARY_PATH="/lib/x86_64-linux-gnu/"
DOCKER_MANIFEST_TEMPLATE="docker/inputs/${APP}.manifest.template"

export COMPOSE_BAKE=true

# Default target
.PHONY: all
all: fmt vet lint test build

# ─────────────────────────────────────────────────────────────────────────────
# Build Targets
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: build
build:
	@echo ">> Building $(APP)..."
	$(GOCMD) build -o bin/$(APP) ./cmd/server

.PHONY: build-cross
build-cross:
	@echo ">> Cross-compiling for linux/amd64..."
	GOOS=linux GOARCH=amd64 $(GOCMD) build -o bin/$(APP)-linux-amd64 ./cmd/server

# ─────────────────────────────────────────────────────────────────────────────
# Run and Test
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: run
run: build
	@echo ">> Running $(APP)..."
	./bin/$(APP)

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
	go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...


# ─────────key-openssl────────────────────────────────────────────────────
# Generate the private key for SGX signing using OpenSSL (exponent 3)
# Usage: make generate-enclave-signing-key
# Requires OpenSSL 1.1.1 or later
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: generate-enclave-signing-key
generate-enclave-signing-key:
	@mkdir -p secrets
	@echo ">> Generating enclave private key at secrets/enclave-key. && \pem using OpenSSL ..."
	@openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -pkeyopt rsa_keygen_pubexp:3 -out secrets/enclave-key.pem
	@chmod 600 secrets/enclave-key.pem
	@echo ">> Enclave signing key generated successfully!"


# ─────────────────────────────────────────────────────────────────────────────
# Generate manifest template
# Usage: make generate-manifest-template
# Requires Gramine to be installed
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: generate-manifest-template
generate-manifest-template:
	chmod +x scripts/generate-manifest-template.sh
	@scripts/generate-manifest-template.sh $(APP) $(DOCKER_MANIFEST_TEMPLATE) $(LD_LIBRARY_PATH)

# ─────────────────────────────────────────────────────────────────────────────
# Extract enclave artifacts
# ─────────────────────────────────────────────────────────────────────────────
# Usage: make extract-enclave-artifacts
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: extract-enclave-artifacts
extract-enclave-artifacts: docker-build
	@echo ">> Getting enclave info..."
	@echo ">> Creating temporary container..."
	@container_id=$$(docker create $(APP)) && \
	echo ">> Container ID: $$container_id" && \
	echo ">> Copying enclave signature file..." && \
	mkdir -p enclave_artifacts && \
	docker cp $$container_id:/app/${APP}.manifest.sgx ./enclave_artifacts/${APP}.manifest.sgx && \
	docker cp $$container_id:/app/${APP}.sig  ./enclave_artifacts/${APP}.sig && \
	docker cp $$container_id:/app/${APP}.metadata.json ./enclave_artifacts/${APP}.metadata.json && \
	echo ">> Removing temporary container..." && \
	docker rm $$container_id && \
	echo ">> Enclave info extracted successfully!"

# ─────────────────────────────────────────────────────────────────────────────
# Native Build and Run
# Usage: make native-build
# Requires Gramine to be installed
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: native-clean
native-clean:
	@echo ">> Cleaning native build artifacts..."
	rm -rf native/outputs/*

.PHONY: native-build
native-build: native-clean
	@echo ">> Building native binary..."
	$(GOCMD) build -o native/outputs/$(APP) ./cmd/server

.PHONY: native-run
native-run: native-build
	@echo ">> Running native application with Gramine..."
	@chmod +x native/run-native.sh
	./native/run-native.sh

# ─────────────────────────────────────────────────────────────────────────────
# Docker Management (Build, Run, Stop, Logs, Status)
# ─────────────────────────────────────────────────────────────────────────────
# Docker compose flags (can be overridden from command line)
DOCKER_FLAGS ?= -d
DOCKER_COMPOSE_FILE ?= docker-compose.yml
DOCKER_SERVICES ?= $(APP)

.PHONY: docker-build
docker-build: generate-manifest-template
	@mkdir -p enclave_artifacts
	DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker compose build $(APP)
	docker tag $(APP):latest $(APP):$(COMMIT)

.PHONY: docker-run
docker-run: docker-build
	@echo ">> Running Docker container..."
	docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES) $(DOCKER_FLAGS)

.PHONY: docker-run-fg
docker-run-fg: docker-build
	docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES)

.PHONY: docker-run-rebuild
docker-run-rebuild: docker-build
	docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES) --build --force-recreate

.PHONY: docker-stop
docker-stop:
	@echo ">> Stopping all Docker containers..."
	docker compose -f $(DOCKER_COMPOSE_FILE) down

.PHONY: docker-logs
docker-logs:
	@echo ">> Showing Docker logs..."
	docker compose -f $(DOCKER_COMPOSE_FILE) logs -f

.PHONY: docker-status
docker-status:
	@echo ">> Docker container status:"
	docker compose -f $(DOCKER_COMPOSE_FILE) ps

# ─────────────────────────────────────────────────────────────────────────────
# Prometheus Monitoring And Alertmanager Setup
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: setup-alertmanager
setup-alertmanager:
	@chmod +x scripts/setup-alertmanager.sh
	@echo ">> Setting up Alertmanager..."
	@scripts/setup-alertmanager.sh

.PHONY: setup-prometheus
setup-prometheus:
	@chmod +x scripts/setup-prometheus.sh
	@echo ">> Setting up Prometheus..."
	@scripts/setup-prometheus.sh

.PHONY: setup-monitoring
setup-monitoring: setup-alertmanager setup-prometheus
	@echo ">> Monitoring setup complete!"

# ─────────────────────────────────────────────────────────────────────────────
# Clean & Help
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: clean
clean:
	@echo ">> Cleaning binaries..."
	rm -rf bin/*

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
	@echo "  build           Compile binary"
	@echo "  build-cross     Build linux/amd64 binary"
	@echo "  run             Build and run locally"
	@echo "  test            Run tests"
	@echo "  fmt             Format code"
	@echo "  vet             Static vetting"
	@echo "  lint            Static analysis (staticcheck)"
	@echo "  clean           Remove binaries"
	@echo
	@echo "Native Targets:"
	@echo "  native-build    Build native binary"
	@echo "  native-run      Build and run native application with Gramine"
	@echo "  native-clean    Clean native build artifacts"
	@echo
	@echo "Docker Targets:"
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
	@echo "  generate-manifest-template   Generate manifest template"
	@echo "  extract-enclave-artifacts    Extract enclave artifacts"
	@echo "  generate-enclave-signing-key Generate enclave signing key (OpenSSL)"
	@echo
	@echo "Utility:"
	@echo "  help            Show this help"
