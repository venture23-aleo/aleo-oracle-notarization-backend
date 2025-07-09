#!/bin/bash

# Nginx Installation Script for Ubuntu
# This script installs and configures nginx on Ubuntu systems

set -e  # Exit immediately on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo "🚀 Installing Nginx on Ubuntu..."

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   print_error "This script should not be run as root"
   exit 1
fi

# Check if nginx is already installed
if command -v nginx &> /dev/null; then
    print_warning "Nginx is already installed!"
    read -p "Do you want to reinstall nginx? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Installation cancelled."
        exit 0
    fi
    print_status "Removing existing nginx installation..."
    sudo apt-get remove --purge nginx nginx-common -y
    sudo apt-get autoremove -y
fi

# Update package index
print_status "Updating package index..."
sudo apt-get update

# Install nginx
print_status "Installing nginx..."
sudo apt-get install -y nginx

# Start and enable nginx service
print_status "Starting and enabling nginx service..."
sudo systemctl start nginx
sudo systemctl enable nginx

# Check nginx status
print_status "Checking nginx status..."
if sudo systemctl is-active --quiet nginx; then
    print_success "Nginx is running successfully!"
else
    print_error "Nginx failed to start!"
    sudo systemctl status nginx
    exit 1
fi

# Configure firewall (if ufw is active)
if sudo ufw status | grep -q "Status: active"; then
    print_status "Configuring firewall for nginx..."
    sudo ufw allow 'Nginx Full'
    print_success "Firewall configured for nginx!"
fi

# Create a basic nginx configuration backup
print_status "Creating backup of default nginx configuration..."
sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup.$(date +%Y%m%d_%H%M%S)

# Test nginx configuration
print_status "Testing nginx configuration..."
if sudo nginx -t; then
    print_success "Nginx configuration is valid!"
else
    print_error "Nginx configuration has errors!"
    exit 1
fi

# Reload nginx to apply any changes
print_status "Reloading nginx..."
sudo systemctl reload nginx

# Display installation information
echo ""
print_success "Nginx installation completed successfully!"
echo ""
echo "📋 Installation Summary:"
echo "  - Nginx version: $(nginx -v 2>&1)"
echo "  - Service status: $(sudo systemctl is-active nginx)"
echo "  - Configuration file: /etc/nginx/nginx.conf"
echo "  - Sites directory: /etc/nginx/sites-available/"
echo "  - Log files: /var/log/nginx/"
echo ""
echo "�� Useful Commands:"
echo "  - Start nginx:     sudo systemctl start nginx"
echo "  - Stop nginx:      sudo systemctl stop nginx"
echo "  - Restart nginx:   sudo systemctl restart nginx"
echo "  - Reload config:   sudo systemctl reload nginx"
echo "  - Check status:    sudo systemctl status nginx"
echo "  - View logs:       sudo tail -f /var/log/nginx/access.log"
echo "  - Test config:     sudo nginx -t"
echo ""
echo "🌐 Default Configuration:"
echo "  - Default site:    /etc/nginx/sites-available/default"
echo "  - Document root:   /var/www/html/"
echo "  - Port:            80 (HTTP)"
echo ""
print_warning "Remember to:"
echo "  - Configure your domain in /etc/nginx/sites-available/"
echo "  - Set up SSL certificates for HTTPS"
echo "  - Update your firewall rules if needed"
echo ""
print_success "Nginx is ready to serve your web applications!"