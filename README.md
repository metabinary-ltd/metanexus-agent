# MetaNexus Agent

Linux agent for MetaNexus infrastructure orchestration platform.

## Building

```bash
# Build for current platform
make build

# Build for Linux x86_64
make build-linux-amd64

# Build for Linux arm64
make build-linux-arm64

# Build for both architectures
make build-all
```

## Installation

```bash
# Build the agent first
make build-linux-amd64  # or build-linux-arm64

# Run installation script
sudo ./scripts/install.sh
```

## Configuration

Create `/etc/metanexus/agent.yaml`:

```yaml
control_plane_url: http://your-control-plane:3000
data_dir: /var/lib/metanexus-agent
telemetry_interval: 30s
```

## Pairing

1. Generate pairing token in control plane UI
2. Run agent with pairing token:
   ```bash
   metanexus-agent -pairing-token <token> -control-plane-url <url>
   ```

## Running

```bash
# As systemd service
sudo systemctl start metanexus-agent

# Manually
metanexus-agent -config /etc/metanexus/agent.yaml
```

## Requirements

- Linux (x86_64 or arm64)
- Docker (optional, for container management actions)
- Network access to control plane

