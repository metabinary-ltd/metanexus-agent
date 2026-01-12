package collect

import (
	"net"
	"os"
	"runtime"
	"strings"
)

func (c *Collector) CollectNetwork() (*TopologyData, error) {
	interfaces := []InterfaceInfo{}
	arpNeighbors := []ARPNeighbor{}

	// Get network interfaces
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip == nil || ip.IsLoopback() || ip.To4() == nil {
					continue
				}

				interfaces = append(interfaces, InterfaceInfo{
					Name:    iface.Name,
					IP:      ip.String(),
					Netmask: addr.String(),
					MAC:     iface.HardwareAddr.String(),
				})
			}
		}
	}

	// Get ARP neighbors (Linux)
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/net/arp"); err == nil {
			lines := strings.Split(string(data), "\n")
			for i, line := range lines {
				if i == 0 || strings.TrimSpace(line) == "" {
					continue // Skip header
				}
				parts := strings.Fields(line)
				if len(parts) >= 6 {
					ip := parts[0]
					mac := parts[3]
					interfaceName := parts[5]

					if mac != "00:00:00:00:00:00" && strings.Contains(mac, ":") {
						arpNeighbors = append(arpNeighbors, ARPNeighbor{
							IP:        ip,
							MAC:       mac,
							Interface: interfaceName,
						})
					}
				}
			}
		}
	}

	return &TopologyData{
		Interfaces:   interfaces,
		ARPNeighbors: arpNeighbors,
		Containers:   []ContainerInfo{}, // Will be populated by Docker collector
	}, nil
}

