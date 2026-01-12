package model

// Generated types from metanexus-spec schemas
// This will be generated from JSON schemas in the future

// Placeholder types matching the telemetry schema
type Telemetry struct {
	AgentID   string                 `json:"agent_id"`
	Timestamp string                 `json:"timestamp"`
	Inventory Inventory              `json:"inventory"`
	Health    Health                 `json:"health"`
	Topology  Topology               `json:"topology"`
}

type Inventory struct {
	Hostname string  `json:"hostname"`
	OS       OS      `json:"os"`
	CPU      CPU     `json:"cpu"`
	Memory   Memory  `json:"memory"`
	Disk     []Disk  `json:"disk"`
}

type OS struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kernel  string `json:"kernel"`
}

type CPU struct {
	Cores int    `json:"cores"`
	Model string `json:"model"`
}

type Memory struct {
	TotalBytes     uint64 `json:"total_bytes"`
	AvailableBytes uint64 `json:"available_bytes,omitempty"`
}

type Disk struct {
	Mount      string `json:"mount"`
	TotalBytes uint64 `json:"total_bytes"`
	UsedBytes  uint64 `json:"used_bytes,omitempty"`
	Filesystem string `json:"filesystem,omitempty"`
}

type Health struct {
	Load              []float64 `json:"load"`
	MemoryUsagePercent float64   `json:"memory_usage_percent"`
	DiskUsagePercent   float64   `json:"disk_usage_percent,omitempty"`
	UptimeSeconds     uint64    `json:"uptime_seconds"`
}

type Topology struct {
	Interfaces   []Interface   `json:"interfaces"`
	ARPNeighbors []ARPNeighbor `json:"arp_neighbors,omitempty"`
	Containers   []Container   `json:"containers"`
}

type Interface struct {
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Netmask string `json:"netmask,omitempty"`
	MAC     string `json:"mac"`
}

type ARPNeighbor struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	Interface string `json:"interface,omitempty"`
}

type Container struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	Status  string   `json:"status"`
	Networks []string `json:"networks,omitempty"`
	IP      string   `json:"ip,omitempty"`
}

