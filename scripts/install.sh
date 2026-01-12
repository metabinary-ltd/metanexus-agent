#!/bin/bash
set -e

# MetaNexus Agent Installation Script

AGENT_BINARY="agent"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
DATA_DIR="/var/lib/metanexus-agent"
CONFIG_DIR="/etc/metanexus"

echo "Installing MetaNexus Agent..."

# Create directories
mkdir -p "$DATA_DIR"
mkdir -p "$CONFIG_DIR"

# Copy binary
if [ -f "./bin/$AGENT_BINARY" ]; then
    cp "./bin/$AGENT_BINARY" "$INSTALL_DIR/metanexus-agent"
    chmod +x "$INSTALL_DIR/metanexus-agent"
    echo "Binary installed to $INSTALL_DIR/metanexus-agent"
else
    echo "Error: Binary not found at ./bin/$AGENT_BINARY"
    echo "Please build the agent first: make build"
    exit 1
fi

# Install systemd service
if [ -f "./scripts/systemd/metanexus-agent.service" ]; then
    cp "./scripts/systemd/metanexus-agent.service" "$SERVICE_DIR/"
    systemctl daemon-reload
    echo "Systemd service installed"
else
    echo "Warning: Systemd service file not found"
fi

echo "Installation complete!"
echo ""
echo "To start the agent:"
echo "  sudo systemctl start metanexus-agent"
echo ""
echo "To enable on boot:"
echo "  sudo systemctl enable metanexus-agent"
echo ""
echo "To check status:"
echo "  sudo systemctl status metanexus-agent"

