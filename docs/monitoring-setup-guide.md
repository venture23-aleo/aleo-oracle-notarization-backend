# Monitoring Setup Guide

## Overview

Setup monitoring for the SGX Go application using Prometheus for metrics collection and Alertmanager for alerting. Works with both Docker and Native deployments. You can use your own `Discord` webhook URL to receive notifications.

## Quick Setup

```bash
# Install and configure all monitoring components
make setup-monitoring
```

This installs and configures:
- Prometheus (metrics collection)
- Alertmanager (alert routing)
- Monitoring for both Docker and Native deployments

## Manual Setup

### Alertmanager Setup
```bash
make setup-alertmanager
```

### Prometheus Setup
```bash
make setup-prometheus
```

## Configuration

### Alertmanager (Discord Notifications)

**Note:** Please replace `YOUR_DISCORD_WEBHOOK_URL` with your own Discord webhook URL.

Located at `/etc/alertmanager/alertmanager.yml`:

```yaml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alertmanager@example.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'discord'

receivers:
  - name: 'discord'
    discord_configs:
      - webhook_url: 'YOUR_DISCORD_WEBHOOK_URL'
```

### Prometheus Configuration
Located at `/etc/prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "/etc/prometheus/alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'sgx-go-app'
    static_configs:
      - targets: ['localhost:8001']
    metrics_path: '/metrics'
```

## Verification

```bash
# Check service status
sudo systemctl status alertmanager
sudo systemctl status prometheus

# Check ports
netstat -tlnp | grep -E '(9090|9093)'
```

## Access

- **Prometheus UI:** http://localhost:9090
- **Alertmanager UI:** http://localhost:9093
- **Application Metrics:** http://localhost:8001/metrics

## Troubleshooting

```bash
# Check logs
sudo journalctl -u prometheus -f
sudo journalctl -u alertmanager -f

# Test metrics endpoint
curl http://localhost:8001/metrics

# Verify configuration
sudo alertmanager --config.file=/etc/alertmanager/alertmanager.yml --check-config
```

## Discord Setup

1. Create a Discord webhook URL in your server
2. Replace `YOUR_DISCORD_WEBHOOK_URL` in the Alertmanager config
3. Restart Alertmanager: `sudo systemctl restart alertmanager`

The monitoring setup automatically works with both Docker and Native deployments. 