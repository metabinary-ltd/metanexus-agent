package collect

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func (c *Collector) CollectInventory() (*InventoryData, error) {
	hostname, _ := os.Hostname()

	// Get OS info
	osName := runtime.GOOS
	osVersion := "unknown"
	kernel := "unknown"

	if runtime.GOOS == "linux" {
		// Try to get OS version from /etc/os-release
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					osVersion = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
					break
				}
			}
		}

		// Get kernel version
		if data, err := os.ReadFile("/proc/version"); err == nil {
			parts := strings.Fields(string(data))
			if len(parts) > 2 {
				kernel = parts[2]
			}
		}
	}

	// Get CPU info
	cpuCores := runtime.NumCPU()
	cpuModel := "unknown"
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "model name") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						cpuModel = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}
	}

	// Get memory info
	memoryTotal := uint64(0)
	memoryAvailable := uint64(0)
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/meminfo"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "MemTotal:") {
					parts := strings.Fields(line)
					if len(parts) > 1 {
						if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
							memoryTotal = val * 1024 // Convert from KB to bytes
						}
					}
				}
				if strings.HasPrefix(line, "MemAvailable:") {
					parts := strings.Fields(line)
					if len(parts) > 1 {
						if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
							memoryAvailable = val * 1024
						}
					}
				}
			}
		}
	}

	// Get disk info
	disks := []DiskInfo{}
	if runtime.GOOS == "linux" {
		cmd := exec.Command("df", "-B1", "--output=target,size,used,fstype")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for i, line := range lines {
				if i == 0 || strings.TrimSpace(line) == "" {
					continue // Skip header
				}
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					mount := parts[0]
					totalStr := parts[1]
					usedStr := parts[2]
					filesystem := parts[3]

					total, _ := strconv.ParseUint(totalStr, 10, 64)
					used, _ := strconv.ParseUint(usedStr, 10, 64)

					disks = append(disks, DiskInfo{
						Mount:      mount,
						TotalBytes: total,
						UsedBytes:  used,
						Filesystem: filesystem,
					})
				}
			}
		}
	}

	return &InventoryData{
		Hostname: hostname,
		OS: OSInfo{
			Name:    osName,
			Version: osVersion,
			Kernel:  kernel,
		},
		CPU: CPUInfo{
			Cores: cpuCores,
			Model: cpuModel,
		},
		Memory: MemoryInfo{
			TotalBytes:     memoryTotal,
			AvailableBytes: memoryAvailable,
		},
		Disk: disks,
	}, nil
}

