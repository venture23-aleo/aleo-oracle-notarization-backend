

MONITORING_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

MONITORING_SCRIPTS_DIR=$(MONITORING_DIR)/scripts

.PHONY: all
all : 
	@echo ">> Please specify a target. See 'make help' for available targets."

# ─────────────────────────────────────────────────────────────────────────────
# Prometheus Monitoring And Alertmanager Setup
# ─────────────────────────────────────────────────────────────────────────────
.PHONY: setup-prometheus setup-alertmanager setup

# Setup Alertmanager
setup-alertmanager:
	@chmod +x $(MONITORING_SCRIPTS_DIR)/setup-alertmanager.sh
	@$(MONITORING_SCRIPTS_DIR)/setup-alertmanager.sh

setup-prometheus:
	@chmod +x $(MONITORING_SCRIPTS_DIR)/setup-prometheus.sh
	@$(MONITORING_SCRIPTS_DIR)/setup-prometheus.sh

.PHONY: setup
setup: setup-prometheus setup-alertmanager 
	@echo ">> Monitoring setup complete!"

.PHONY: help
help:
	@echo "Monitoring Targets:"
	@echo "  setup           Complete monitoring setup"
	@echo "  setup-prometheus Setup Prometheus"
	@echo "  setup-alertmanager Setup Alertmanager"