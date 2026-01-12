package collect

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func (c *Collector) CollectHealth() (*HealthData, error) {
	load := []float64{0, 0, 0}
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/loadavg"); err == nil {
			parts := strings.Fields(string(data))
			if len(parts) >= 3 {
				load[0], _ = strconv.ParseFloat(parts[0], 64)
				load[1], _ = strconv.ParseFloat(parts[1], 64)
				load[2], _ = strconv.ParseFloat(parts[2], 64)
			}
		}
	}

	// Get memory usage
	memoryUsagePercent := float64(0)
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/meminfo"); err == nil {
			var memTotal, memAvailable uint64
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "MemTotal:") {
					parts := strings.Fields(line)
					if len(parts) > 1 {
						if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
							memTotal = val * 1024
						}
					}
				}
				if strings.HasPrefix(line, "MemAvailable:") {
					parts := strings.Fields(line)
					if len(parts) > 1 {
						if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
							memAvailable = val * 1024
						}
					}
				}
			}
			if memTotal > 0 {
				memoryUsagePercent = float64(memTotal-memAvailable) / float64(memTotal) * 100
			}
		}
	}

	// Get disk usage (primary mount)
	diskUsagePercent := float64(0)
	if runtime.GOOS == "linux" {
		// Use df to get root filesystem usage
		// This is simplified - in production, would parse df output properly
	}

	// Get uptime
	uptimeSeconds := uint64(0)
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/uptime"); err == nil {
			parts := strings.Fields(string(data))
			if len(parts) > 0 {
				if uptime, err := strconv.ParseFloat(parts[0], 64); err == nil {
					uptimeSeconds = uint64(uptime)
				}
			}
		}
	}

	return &HealthData{
		Load:              load,
		MemoryUsagePercent: memoryUsagePercent,
		DiskUsagePercent:   diskUsagePercent,
		UptimeSeconds:     uptimeSeconds,
	}, nil
}

