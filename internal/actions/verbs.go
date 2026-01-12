package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/metabinary-ltd/metanexus-agent/internal/collect"
	"github.com/metabinary-ltd/metanexus-agent/internal/identity"
)

type UpdateContainerVerb struct {
	dockerClient *client.Client
}

func NewUpdateContainerVerb() (*UpdateContainerVerb, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker not available: %w", err)
	}

	return &UpdateContainerVerb{dockerClient: cli}, nil
}

func (v *UpdateContainerVerb) Execute(params map[string]interface{}) (map[string]interface{}, error) {
	containerName, ok := params["container_name"].(string)
	if !ok {
		return nil, fmt.Errorf("container_name parameter required")
	}

	pullLatest := true
	if val, ok := params["pull_latest"].(bool); ok {
		pullLatest = val
	}

	restart := true
	if val, ok := params["restart"].(bool); ok {
		restart = val
	}

	ctx := context.Background()

	// Find container
	containers, err := v.dockerClient.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var containerID string
	for _, c := range containers {
		for _, name := range c.Names {
			if name[1:] == containerName { // Remove leading /
				containerID = c.ID
				break
			}
		}
		if containerID != "" {
			break
		}
	}

	if containerID == "" {
		return nil, fmt.Errorf("container not found: %s", containerName)
	}

	// Get container details
	containerInfo, err := v.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	imageName := containerInfo.Config.Image

	// Pull latest image if requested
	if pullLatest {
		_, err := v.dockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to pull image: %w", err)
		}
	}

	// Stop container if running
	if containerInfo.State.Running {
		if err := v.dockerClient.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
			return nil, fmt.Errorf("failed to stop container: %w", err)
		}
	}

	// Remove old container
	if err := v.dockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{}); err != nil {
		return nil, fmt.Errorf("failed to remove container: %w", err)
	}

	// Create new container with same config
	createResp, err := v.dockerClient.ContainerCreate(
		ctx,
		containerInfo.Config,
		containerInfo.HostConfig,
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container if restart requested
	if restart {
		if err := v.dockerClient.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{}); err != nil {
			return nil, fmt.Errorf("failed to start container: %w", err)
		}
	}

	return map[string]interface{}{
		"status":         "completed",
		"container_id":   createResp.ID,
		"container_name": containerName,
		"image":          imageName,
	}, nil
}

type HealthSnapshotVerb struct {
	collector *collect.Collector
}

func NewHealthSnapshotVerb() *HealthSnapshotVerb {
	return &HealthSnapshotVerb{
		collector: collect.New(),
	}
}

func (v *HealthSnapshotVerb) Execute(params map[string]interface{}) (map[string]interface{}, error) {
	includeDisk := true
	if val, ok := params["include_disk"].(bool); ok {
		includeDisk = val
	}

	includeNetwork := true
	if val, ok := params["include_network"].(bool); ok {
		includeNetwork = val
	}

	result := map[string]interface{}{
		"status":    "completed",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}

	// Collect health
	health, err := v.collector.CollectHealth()
	if err == nil {
		result["health"] = map[string]interface{}{
			"load":                 health.Load,
			"memory_usage_percent": health.MemoryUsagePercent,
			"disk_usage_percent":   health.DiskUsagePercent,
			"uptime_seconds":       health.UptimeSeconds,
		}
	}

	// Collect disk info if requested
	if includeDisk {
		inventory, err := v.collector.CollectInventory()
		if err == nil {
			result["disk"] = inventory.Disk
		}
	}

	// Collect network info if requested
	if includeNetwork {
		topology, err := v.collector.CollectTopology()
		if err == nil {
			result["network"] = map[string]interface{}{
				"interfaces": topology.Interfaces,
			}
		}
	}

	return result, nil
}

type RotateIdentityVerb struct {
	identityMgr *identity.Manager
}

func NewRotateIdentityVerb(identityMgr *identity.Manager) *RotateIdentityVerb {
	return &RotateIdentityVerb{
		identityMgr: identityMgr,
	}
}

func (v *RotateIdentityVerb) Execute(params map[string]interface{}) (map[string]interface{}, error) {
	// Generate new identity - need to access unexported method
	// For now, return error indicating re-pairing needed
	return nil, fmt.Errorf("identity rotation requires re-pairing - not yet implemented")

	return map[string]interface{}{
		"status":  "completed",
		"message": "New identity generated. Re-pairing required.",
	}, nil
}
