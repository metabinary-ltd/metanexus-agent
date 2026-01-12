# MetaNexus Agent Capabilities

## Overview

Agents execute actions based on their capabilities. Capabilities define what an agent is allowed to do.

## Available Capabilities

### docker
Allows the agent to:
- List Docker containers
- Update Docker containers
- Inspect container status
- Manage container networks

**Required for actions:**
- `update-container`

### system
Allows the agent to:
- Collect system inventory
- Run health checks
- Monitor system resources

**Required for actions:**
- `health-snapshot`

### network
Allows the agent to:
- Inspect network interfaces
- Discover ARP neighbors
- Analyze network topology

### identity
Allows the agent to:
- Rotate certificates
- Manage agent identity

**Required for actions:**
- `rotate-identity`

## Capability Assignment

Capabilities are assigned during agent registration and can be updated by the control plane. Agents report their capabilities in telemetry.

## Action Execution

Before executing an action, agents verify:
1. Action is in the signed catalog
2. Agent has required capabilities
3. Action signature is valid
4. Action has not expired

Actions without required capabilities are rejected.

