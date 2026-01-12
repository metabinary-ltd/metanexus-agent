package collect

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func (c *Collector) CollectDocker() ([]ContainerInfo, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		// Docker not available
		return []ContainerInfo{}, nil
	}
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		// Docker not available or error
		return []ContainerInfo{}, nil
	}

	result := []ContainerInfo{}
	for _, container := range containers {
		containerInfo := ContainerInfo{
			ID:     container.ID[:12], // Short ID
			Name:   container.Names[0][1:], // Remove leading /
			Image:  container.Image,
			Status:  container.Status,
			Networks: []string{},
		}

		// Get container details for IP
		details, err := cli.ContainerInspect(context.Background(), container.ID)
		if err == nil {
			// Get IP from first network
			for _, network := range details.NetworkSettings.Networks {
				if network.IPAddress != "" {
					containerInfo.IP = network.IPAddress
					break
				}
			}

			// Get network names
			for netName := range details.NetworkSettings.Networks {
				containerInfo.Networks = append(containerInfo.Networks, netName)
			}
		}

		// Map status
		switch {
		case strings.Contains(container.Status, "Up"):
			containerInfo.Status = "running"
		case strings.Contains(container.Status, "Exited"):
			containerInfo.Status = "stopped"
		default:
			containerInfo.Status = "stopped"
		}

		result = append(result, containerInfo)
	}

	return result, nil
}

