package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/metanexus/metanexus-agent/internal/collect"
	"github.com/metanexus/metanexus-agent/pkg/model"
)

func (a *Agent) sendTelemetry(ctx context.Context) error {
	data, err := a.collector.CollectAll()
	if err != nil {
		return fmt.Errorf("failed to collect telemetry: %w", err)
	}

	// Get agent ID
	agentID, err := a.identity.GetAgentID()
	if err != nil {
		return fmt.Errorf("failed to get agent ID: %w", err)
	}

	// Convert to model format
	telemetry := &model.Telemetry{
		AgentID:   agentID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Inventory: convertInventory(data.Inventory),
		Health:    convertHealth(data.Health),
		Topology:  convertTopology(data.Topology),
	}

	// Send via transport
	jsonData, err := json.Marshal(telemetry)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry: %w", err)
	}

	resp, err := a.transport.Post(fmt.Sprintf("/api/v1/agents/%s/telemetry", agentID), telemetry)
	if err != nil {
		return fmt.Errorf("failed to send telemetry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telemetry send failed with status: %d", resp.StatusCode)
	}

	return nil
}

func convertInventory(inv collect.InventoryData) model.Inventory {
	disks := make([]model.Disk, len(inv.Disk))
	for i, d := range inv.Disk {
		disks[i] = model.Disk{
			Mount:      d.Mount,
			TotalBytes: d.TotalBytes,
			UsedBytes:  d.UsedBytes,
			Filesystem: d.Filesystem,
		}
	}

	return model.Inventory{
		Hostname: inv.Hostname,
		OS: model.OS{
			Name:    inv.OS.Name,
			Version: inv.OS.Version,
			Kernel:  inv.OS.Kernel,
		},
		CPU: model.CPU{
			Cores: inv.CPU.Cores,
			Model: inv.CPU.Model,
		},
		Memory: model.Memory{
			TotalBytes:     inv.Memory.TotalBytes,
			AvailableBytes: inv.Memory.AvailableBytes,
		},
		Disk: disks,
	}
}

func convertHealth(health collect.HealthData) model.Health {
	return model.Health{
		Load:              health.Load,
		MemoryUsagePercent: health.MemoryUsagePercent,
		DiskUsagePercent:   health.DiskUsagePercent,
		UptimeSeconds:     health.UptimeSeconds,
	}
}

func convertTopology(topology collect.TopologyData) model.Topology {
	interfaces := make([]model.Interface, len(topology.Interfaces))
	for i, iface := range topology.Interfaces {
		interfaces[i] = model.Interface{
			Name:    iface.Name,
			IP:      iface.IP,
			Netmask: iface.Netmask,
			MAC:     iface.MAC,
		}
	}

	arpNeighbors := make([]model.ARPNeighbor, len(topology.ARPNeighbors))
	for i, neighbor := range topology.ARPNeighbors {
		arpNeighbors[i] = model.ARPNeighbor{
			IP:        neighbor.IP,
			MAC:       neighbor.MAC,
			Interface: neighbor.Interface,
		}
	}

	containers := make([]model.Container, len(topology.Containers))
	for i, container := range topology.Containers {
		containers[i] = model.Container{
			ID:      container.ID,
			Name:    container.Name,
			Image:   container.Image,
			Status:  container.Status,
			Networks: container.Networks,
			IP:      container.IP,
		}
	}

	return model.Topology{
		Interfaces:   interfaces,
		ARPNeighbors: arpNeighbors,
		Containers:   containers,
	}
}

