package collect

import (
	"encoding/json"
	"fmt"
	"time"
)

type Collector struct{}

func New() *Collector {
	return &Collector{}
}

type TelemetryData struct {
	Inventory InventoryData
	Health    HealthData
	Topology  TopologyData
}

type InventoryData struct {
	Hostname string
	OS       OSInfo
	CPU      CPUInfo
	Memory   MemoryInfo
	Disk     []DiskInfo
}

type OSInfo struct {
	Name    string
	Version string
	Kernel  string
}

type CPUInfo struct {
	Cores int
	Model string
}

type MemoryInfo struct {
	TotalBytes     uint64
	AvailableBytes uint64
}

type DiskInfo struct {
	Mount      string
	TotalBytes uint64
	UsedBytes  uint64
	Filesystem string
}

type HealthData struct {
	Load              []float64
	MemoryUsagePercent float64
	DiskUsagePercent   float64
	UptimeSeconds     uint64
}

type TopologyData struct {
	Interfaces   []InterfaceInfo
	ARPNeighbors []ARPNeighbor
	Containers   []ContainerInfo
}

type InterfaceInfo struct {
	Name    string
	IP      string
	Netmask string
	MAC     string
}

type ARPNeighbor struct {
	IP        string
	MAC       string
	Interface string
}

type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Networks []string
	IP      string
}

func (c *Collector) CollectAll() (*TelemetryData, error) {
	inventory, err := c.CollectInventory()
	if err != nil {
		return nil, fmt.Errorf("failed to collect inventory: %w", err)
	}

	health, err := c.CollectHealth()
	if err != nil {
		return nil, fmt.Errorf("failed to collect health: %w", err)
	}

	topology, err := c.CollectTopology()
	if err != nil {
		return nil, fmt.Errorf("failed to collect topology: %w", err)
	}

	return &TelemetryData{
		Inventory: *inventory,
		Health:    *health,
		Topology:  *topology,
	}, nil
}

// CollectInventory is now implemented in inventory.go

// CollectHealth is now implemented in health.go

func (c *Collector) CollectTopology() (*TopologyData, error) {
	// Collect network info
	networkData, err := c.CollectNetwork()
	if err != nil {
		return nil, err
	}

	// Collect Docker containers
	containers, err := c.CollectDocker()
	if err != nil {
		// Docker not available, continue without containers
		containers = []ContainerInfo{}
	}

	return &TopologyData{
		Interfaces:   networkData.Interfaces,
		ARPNeighbors: networkData.ARPNeighbors,
		Containers:   containers,
	}, nil
}

