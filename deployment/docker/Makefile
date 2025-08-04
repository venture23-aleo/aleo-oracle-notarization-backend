DOCKER_DEPLOYMENT_DIR := $(shell cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
DOCKER_SCRIPTS_DIR=$(DOCKER_DEPLOYMENT_DIR)/scripts

DEPLOYMENT_DIR = $(DOCKER_DEPLOYMENT_DIR)/..

export SECRETS_DIR = $(DEPLOYMENT_DIR)/secrets
export ENCLAVE_SIGNING_KEY_FILE ?= "$(SECRETS_DIR)/enclave-signing-key.pem"

export APP ?= aleo-oracle-notarization-backend

DOCKER_COMPOSE_FILE ?= "$(DOCKER_DEPLOYMENT_DIR)/docker-compose.yml"
DOCKER_FLAGS ?= -d
DOCKER_SERVICES ?= $(APP)
COMMIT := $(shell git rev-parse HEAD)

.PHONY: all
all : 
	@echo ">> Please specify a target. See 'make help' for available targets."

.PHONY: install-docker generate-manifest-template generate-enclave-signing-key docker-setup docker-build docker-run docker-run-fg docker-run-rebuild docker-stop docker-logs docker-status extract-enclave-artifacts

# Generate enclave signing key
generate-enclave-signing-key:
	@echo ">> Generating enclave signing key..."
	@chmod +x $(SECRETS_DIR)/generate-enclave-signing-key.sh
	@$(SECRETS_DIR)/generate-enclave-signing-key.sh

# Generate manifest template
generate-manifest-template:
	@chmod +x $(DOCKER_SCRIPTS_DIR)/generate-manifest-template.sh
	@$(DOCKER_SCRIPTS_DIR)/generate-manifest-template.sh

# Extract enclave artifacts
extract-enclave-artifacts:
	@chmod +x $(DOCKER_SCRIPTS_DIR)/extract-enclave-artifacts.sh
	@$(DOCKER_SCRIPTS_DIR)/extract-enclave-artifacts.sh $(APP)


# Install Docker
install-docker:
	@chmod +x $(DOCKER_SCRIPTS_DIR)/install-docker.sh
	@$(DOCKER_SCRIPTS_DIR)/install-docker.sh

# Setup
docker-setup: install-docker generate-enclave-signing-key
	@echo ">> Docker deployment setup completed successfully!"

# Build
docker-build: generate-manifest-template
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_DEPLOYMENT_INPUTS_DIR="$(DOCKER_DEPLOYMENT_INPUTS_DIR)" ENCLAVE_SIGNING_KEY_FILE="$(ENCLAVE_SIGNING_KEY_FILE)" docker compose -f $(DOCKER_COMPOSE_FILE) build $(APP)
	@docker tag $(APP):latest $(APP):$(COMMIT)

# Run
docker-run: docker-build
	@echo ">> Running Docker container..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES) $(DOCKER_FLAGS)

# Run in foreground
docker-run-fg: docker-build
	@echo ">> Running Docker container in foreground..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES)

# Run in foreground with rebuild
docker-run-rebuild: docker-build
	@echo ">> Running Docker container in foreground with rebuild..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) up $(DOCKER_SERVICES) --build --force-recreate

# Stop
docker-stop:
	@echo ">> Stopping all Docker containers..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) down

# Show logs
docker-logs:
	@echo ">> Showing Docker logs..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) logs -f

# Show container status
docker-status:
	@echo ">> Docker container status:"
	@docker compose -f $(DOCKER_COMPOSE_FILE) ps

.PHONY: help
help:
	@echo "Docker Deployment Targets:"
	@echo "  docker-setup           Complete Docker environment setup (recommended)"
	@echo "  install-docker  Install Docker and Docker Compose"
	@echo "  docker-build           Build Docker image"
	@echo "  docker-run           Build and run Docker container (detached mode)"
	@echo "  docker-run-fg        Build and run Docker container (foreground mode)"
	@echo "  docker-run-rebuild   Build and run Docker container (force rebuild)"
	@echo "  docker-stop            Stop all Docker containers"
	@echo "  docker-logs            Show Docker logs"
	@echo "  docker-status          Show Docker container status"
	@echo "  extract-enclave-artifacts    Extract enclave artifacts for Docker"

