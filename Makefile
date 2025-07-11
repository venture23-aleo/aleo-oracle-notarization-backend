
# Makefile for Aleo Oracle Notarizer Development Environment

# Pointing to env. 
# If .env file does not exist, create it
ifeq ($(wildcard .env),)
    $(shell touch .env)
endif

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
MANIFEST_TEMPLATE="build/inputs/${APP}.manifest.template"
SGX_CONF_FILE="build/inputs/sgx_default_qcnl.conf"

export COMPOSE_BAKE=true
export APP := aleo-oracle-notarization-backend

ifeq ($(PORT),)
	PORT := 8000
endif

ifeq ($(METRICS_PORT),)
	METRICS_PORT := 8001
endif

export METRICS_PORT

export PORT
export WHITELISTED_DOMAINS

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

# ─────────────────────────────────────────────────────────────────────────────
# Generate manifest template
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: generate-manifest-template
generate-manifest-template:
	chmod +x scripts/generate-manifest-template.sh
	@scripts/generate-manifest-template.sh $(APP) $(MANIFEST_TEMPLATE) $(LD_LIBRARY_PATH)

# ─────────────────────────────────────────────────────────────────────────────
# Get enclave info
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: get-enclave-info
get-enclave-info: docker-build
	docker compose run --volume "$(shell pwd)/enclave_info.json/:/app/enclave_info.json" --entrypoint /bin/bash --rm $(APP) -c "gramine-sgx-sigstruct-view /app/${APP}.sig -v --output-format json > /app/enclave_info.json"

# ─────────────────────────────────────────────────────────────────────────────
# Docker
# ─────────────────────────────────────────────────────────────────────────────
# Docker compose flags (can be overridden from command line)
DOCKER_FLAGS ?= -d
DOCKER_COMPOSE_FILE ?= docker-compose.yml
DOCKER_SERVICES ?= $(APP)

.PHONY: docker-build
docker-build: generate-manifest-template
	DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker compose build $(APP)
	docker tag $(APP):latest $(APP):$(COMMIT)

.PHONY: docker-run
docker-run: docker-build
	docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES) $(DOCKER_FLAGS)

.PHONY: docker-run-fg
docker-run-fg: docker-build
	docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES)

.PHONY: docker-run-rebuild
docker-run-rebuild: docker-build
	docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES) --build --force-recreate

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
# Docker Management
# ─────────────────────────────────────────────────────────────────────────────
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
# Clean & Help
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: clean
clean:
	@echo ">> Cleaning binaries..."
	rm -rf bin/*

.PHONY: gen-key
# ─────────────────────────────────────────────────────────────────────────────
# Generate the enclave private key for SGX signing
# Usage: make gen-key
# Requires gramine-sgx-gen-private-key to be installed
# ─────────────────────────────────────────────────────────────────────────────
gen-key:
	@mkdir -p secrets
	@echo ">> Generating enclave private key at secrets/enclave-key.pem ..."
	rm -f secrets/enclave-key.pem
	gramine-sgx-gen-private-key secrets/enclave-key.pem

.PHONY: gen-key-openssl
# ─────────────────────────────────────────────────────────────────────────────
# Generate the enclave private key for SGX signing using OpenSSL (exponent 3)
# Usage: make gen-key-openssl
# Requires OpenSSL 1.1.1 or later
# ─────────────────────────────────────────────────────────────────────────────
gen-key-openssl:
	@mkdir -p secrets
	@echo ">> Generating enclave private key at secrets/enclave-key.pem using OpenSSL ..."
	openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -pkeyopt rsa_keygen_pubexp:3 -out secrets/enclave-key.pem

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
	@echo "  generate-manifest-template  Generate manifest template"
	@echo "  get-enclave-info            Get enclave info"
	@echo "  gen-key                     Generate enclave private key (Gramine tool)"
	@echo "  gen-key-openssl             Generate enclave private key (OpenSSL)"
	@echo
	@echo "Utility:"
	@echo "  help            Show this help"
