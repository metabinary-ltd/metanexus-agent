# MetaNexus Agent Security

## Overview

The MetaNexus agent uses mutual TLS (mTLS) for secure communication with the control plane. This document outlines the security model and best practices.

## Identity and Certificates

- Agents generate RSA 2048-bit key pairs on first run
- Certificate signing requests (CSRs) are sent to the control plane during pairing
- Control plane signs certificates using its internal CA
- Certificates are short-lived (90 days default) and should be rotated regularly

## Pairing Process

1. Admin generates pairing token in control plane UI
2. Agent is configured with pairing token and control plane URL
3. Agent generates keypair and CSR
4. Agent sends CSR to control plane with pairing token
5. Control plane validates token and signs certificate
6. Agent stores certificate and CA certificate
7. All subsequent communication uses mTLS

## Communication Security

- All API communication uses HTTPS with mTLS
- Agent certificate must be valid and signed by control plane CA
- Control plane verifies agent certificate on each request
- Telemetry data is encrypted in transit

## Best Practices

1. **Secure Storage**: Agent private keys are stored with 0600 permissions
2. **Certificate Rotation**: Implement certificate rotation before expiration
3. **Network Security**: Run agent on trusted networks or use VPN
4. **Access Control**: Limit filesystem access to agent data directory
5. **Logging**: Monitor agent logs for suspicious activity

## Capabilities

Agents execute actions based on capabilities:
- `docker`: Can manage Docker containers
- `system`: Can run system health checks
- `network`: Can inspect network configuration
- `identity`: Can rotate certificates

Actions are cryptographically signed by the control plane and verified by agents before execution.

