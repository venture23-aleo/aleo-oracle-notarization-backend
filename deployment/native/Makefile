# ────────────────────────────────────────────────────────────────────────────
# Native Deployment Scripts
# ────────────────────────────────────────────────────────────────────────────
export NATIVE_DEPLOYMENT_DIR=$(shell cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
export NATIVE_DEPLOYMENT_SCRIPTS_DIR=$(NATIVE_DEPLOYMENT_DIR)/scripts

export DEPLOYMENT_DIR = $(NATIVE_DEPLOYMENT_DIR)/..
export SECRETS_DIR = $(DEPLOYMENT_DIR)/secrets
export ENCLAVE_SIGNING_KEY_FILE ?= "$(SECRETS_DIR)/enclave-signing-key.pem"

.PHONY: install-sgx-dcap-aesm install-gramine install-go generate-manifest-template generate-enclave-signing-key native-setup native-run native-clean

# Generate enclave signing key
generate-enclave-signing-key:
	@chmod +x $(SECRETS_DIR)/generate-enclave-signing-key.sh
	@$(SECRETS_DIR)/generate-enclave-signing-key.sh

# Generate manifest template
generate-manifest-template:
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/generate-manifest-template.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/generate-manifest-template.sh

# Install SGX DCAP driver and AESM service
install-sgx-dcap-aesm:
	@echo ">> Installing SGX DCAP driver and AESM service..."
	@echo "NATIVE_DEPLOYMENT_SCRIPTS_DIR: $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)"
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-sgx-dcap-aesm.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-sgx-dcap-aesm.sh

# Install Gramine
install-gramine:
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-gramine.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-gramine.sh

# Install Go
install-go:
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-go.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/install-go.sh

# Setup
native-setup: install-sgx-dcap-aesm install-gramine install-go generate-enclave-signing-key
	@echo ">> Native setup completed successfully!"

# Run Gramine enclave
native-run: generate-manifest-template
	@chmod +x $(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/run.sh
	@$(NATIVE_DEPLOYMENT_SCRIPTS_DIR)/run.sh

# Clean native build artifacts
native-clean:
	@echo ">> Cleaning native build artifacts..."
	@rm -rf $(NATIVE_DEPLOYMENT_ENCLAVE_ARTIFACTS_DIR)/*


.PHONY: help
help:
	@echo "Native Deployment Targets:"
	@echo "  install-sgx-dcap-aesm Install Intel SGX and DCAP drivers"
	@echo "  install-gramine Install Gramine framework"
	@echo "  install-go      Install Go 1.24.4"
	@echo "  generate-manifest-template Generate Gramine manifest template"
	@echo "  generate-enclave-signing-key Generate enclave signing key"
	@echo "  native-setup           Complete native environment setup (recommended)"
	@echo "  native-run           Build and run native application with Gramine"
	@echo "  native-clean           Clean native build artifacts"