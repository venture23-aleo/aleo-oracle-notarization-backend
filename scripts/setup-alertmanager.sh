#!/usr/bin/env bash

set -e

ALERT_VERSION="0.27.0"
ALERT_USER="alertmanager"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/alertmanager"
DATA_DIR="/var/lib/alertmanager"
SERVICE_FILE="/etc/systemd/system/alertmanager.service"

echo "🔧 Installing Alertmanager v${ALERT_VERSION}..."

# Create a system user and group
sudo groupadd --system ${ALERT_USER} || true
sudo useradd -s /sbin/nologin --system -g ${ALERT_USER} ${ALERT_USER} || true

# Create required directories
sudo mkdir -p ${CONFIG_DIR} ${DATA_DIR}
sudo chown -R ${ALERT_USER}:${ALERT_USER} ${CONFIG_DIR} ${DATA_DIR}

sudo cp $PWD/scripts/inputs/alertmanager.yml ${CONFIG_DIR}/alertmanager.yml

# Download and extract Alertmanager
cd /tmp
wget -q "https://github.com/prometheus/alertmanager/releases/download/v${ALERT_VERSION}/alertmanager-${ALERT_VERSION}.linux-amd64.tar.gz"
tar -xzf alertmanager-${ALERT_VERSION}.linux-amd64.tar.gz

# Move binaries
cd alertmanager-${ALERT_VERSION}.linux-amd64
sudo mv alertmanager amtool ${INSTALL_DIR}/
sudo chown ${ALERT_USER}:${ALERT_USER} ${INSTALL_DIR}/alertmanager ${INSTALL_DIR}/amtool

sudo chown ${ALERT_USER}:${ALERT_USER} ${CONFIG_DIR}/alertmanager.yml

# Create systemd service file
cat <<EOF | sudo tee ${SERVICE_FILE}
[Unit]
Description=Alertmanager
Wants=network-online.target
After=network-online.target

[Service]
User=${ALERT_USER}
Group=${ALERT_USER}
Type=simple
ExecStart=${INSTALL_DIR}/alertmanager \\
  --config.file=${CONFIG_DIR}/alertmanager.yml \\
  --storage.path=${DATA_DIR}

[Install]
WantedBy=multi-user.target
EOF

# Enable and start Alertmanager
sudo systemctl daemon-reload
sudo systemctl enable alertmanager
sudo systemctl start alertmanager

echo "✅ Alertmanager installed and running on port 9093"
echo "🔗 Access it at: http://localhost:9093"
