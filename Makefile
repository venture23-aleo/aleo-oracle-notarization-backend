# Makefile for Aleo Oracle Notarizer Development Environment

# Pointing to dev env. 
include .env
export

# ─────────────────────────────────────────────────────────────────────────────
# Variables
# ─────────────────────────────────────────────────────────────────────────────
SHELL := /bin/bash
export VERSION := $(shell git describe --tags --always --dirty)
export COMMIT := $(shell git rev-parse HEAD)

LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)"
GOCMD := go
LD_LIBRARY_PATH="/lib/x86_64-linux-gnu/"
MANIFEST_TEMPLATE="build/inputs/${APP}.manifest.template"
SGX_CONF_FILE="build/inputs/sgx_default_qcnl.conf"

# Default target
.PHONY: all
all: fmt vet lint test build

# ─────────────────────────────────────────────────────────────────────────────
# Build Targets
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: build
build:
	@echo ">> Building $(APP)..."
	$(GOCMD) build $(LDFLAGS) -o bin/$(APP) ./cmd/server

.PHONY: build-cross
build-cross:
	@echo ">> Cross-compiling for linux/amd64..."
	GOOS=linux GOARCH=amd64 $(GOCMD) build $(LDFLAGS) -o bin/$(APP)-linux-amd64 ./cmd/server

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
# Quality Checks
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: fmt
fmt:
	@echo ">> Formatting code..."
	$(GOCMD) fmt ./...

.PHONY: vet
vet:
	@echo ">> Vetting code..."
	$(GOCMD) vet ./...

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
# Docker
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: docker-build
docker-build: generate-manifest-template
	docker compose build $(APP)
	docker tag $(APP):latest $(APP):$(VERSION)-$(COMMIT)
	
.PHONY: docker-run
docker-run: docker-build
	docker compose up $(APP)

# ─────────────────────────────────────────────────────────────────────────────
# Clean & Help
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: clean
clean:
	@echo ">> Cleaning binaries..."
	rm -rf bin/*

.PHONY: help
help:
	@echo "Usage:"
	@echo "  make [target]"
	@echo
	@echo "Targets:"
	@echo "  all           Run fmt, vet, lint, test, build"
	@echo "  build         Compile binary"
	@echo "  build-cross   Build linux/amd64 binary"
	@echo "  run           Build and run locally"
	@echo "  test          Run tests"
	@echo "  fmt           Format code"
	@echo "  vet           Static vetting"
	@echo "  lint          Static analysis (staticcheck)"
	@echo "  docker-build  Build Docker image"
	@echo "  docker-run    Run Docker container"
	@echo "  clean         Remove binaries"
	@echo "  help          Show this help"
