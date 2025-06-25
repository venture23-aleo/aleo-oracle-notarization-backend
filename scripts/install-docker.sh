#!/bin/bash

set -e  # Exit immediately on error

echo "[+] Updating package index..."
sudo apt-get update

echo "[+] Installing prerequisites..."
sudo apt-get install -y ca-certificates curl gnupg lsb-release

echo "[+] Creating keyring directory..."
sudo install -m 0755 -d /etc/apt/keyrings

echo "[+] Fetching Docker's GPG key..."
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc

echo "[+] Setting GPG key permissions..."
sudo chmod a+r /etc/apt/keyrings/docker.asc

echo "[+] Adding Docker repository to APT sources..."
ARCH=$(dpkg --print-architecture)
CODENAME=$(source /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")

echo "deb [arch=$ARCH signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $CODENAME stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

echo "[+] Updating package index (with Docker repo)..."
sudo apt-get update

echo "[+] Installing Docker packages..."
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

echo "[✓] Docker installation complete."
echo "[✓] Verifying Docker version:"
docker --version
