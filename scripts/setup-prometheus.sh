#!/usr/bin/env bash
set -e

PROM_VERSION="2.43.0"
PROM_USER="prometheus"
PROM_GROUP="prometheus"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/prometheus"
DATA_DIR="/var/lib/prometheus"
SERVICE_FILE="/etc/systemd/system/prometheus.service"

echo "⏳ Updating package list..."
sudo apt update

echo "👤 Creating Prometheus user and group..."
sudo groupadd --system ${PROM_GROUP} || true
sudo useradd -s /sbin/nologin --system -g ${PROM_GROUP} ${PROM_USER} || true

echo "📁 Creating directories..."
sudo mkdir -p ${CONFIG_DIR} ${DATA_DIR}
sudo chown ${PROM_USER}:${PROM_GROUP} ${CONFIG_DIR} ${DATA_DIR}

echo "⬇️  Downloading Prometheus ${PROM_VERSION}..."
cd /tmp
wget -q "https://github.com/prometheus/prometheus/releases/download/v${PROM_VERSION}/prometheus-${PROM_VERSION}.linux-amd64.tar.gz"
tar xzf prometheus-${PROM_VERSION}.linux-amd64.tar.gz

cd prometheus-${PROM_VERSION}.linux-amd64
echo "🚚 Installing binaries..."
sudo mv prometheus promtool ${INSTALL_DIR}/
sudo chown ${PROM_USER}:${PROM_GROUP} ${INSTALL_DIR}/prometheus ${INSTALL_DIR}/promtool

echo "⚙️  Installing config and console files..."
sudo mv consoles console_libraries prometheus.yml ${CONFIG_DIR}/
sudo chown -R ${PROM_USER}:${PROM_GROUP} ${CONFIG_DIR}
sudo chown -R ${PROM_USER}:${PROM_GROUP} ${DATA_DIR}

echo "🛠️  Creating systemd service file..."
sudo tee ${SERVICE_FILE} > /dev/null <<EOF
[Unit]
Description=Prometheus Monitoring
Wants=network-online.target
After=network-online.target

[Service]
User=${PROM_USER}
Group=${PROM_GROUP}
Type=simple
ExecStart=${INSTALL_DIR}/prometheus \\
  --config.file=${CONFIG_DIR}/prometheus.yml \\
  --storage.tsdb.path=${DATA_DIR}/ \\
  --web.console.templates=${CONFIG_DIR}/consoles \\
  --web.console.libraries=${CONFIG_DIR}/console_libraries

[Install]
WantedBy=multi-user.target
EOF

sudo cp ./scripts/inputs/prometheus.yml ${CONFIG_DIR}/prometheus.yml
sudo cp ./scripts/inputs/alerts.yml ${CONFIG_DIR}/alerts.yml

echo "🔄 Reloading systemd configuration..."
sudo systemctl daemon-reload

echo "▶️ Enabling and starting Prometheus..."
sudo systemctl enable prometheus
sudo systemctl start prometheus

echo "✅ Prometheus install complete. Status:"
sudo systemctl status prometheus --no-pager

echo "🌐 Prometheus UI: http://localhost:9090"
echo "🔐 Remember to secure access using firewall, reverse proxy, or VPN."
