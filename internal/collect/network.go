package collect

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
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

			// Collect all addresses (IPv4 and IPv6)
			for _, addr := range addrs {
				var ip net.IP
				var network string
				var isIPv6 bool

				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
					network = v.String()
					isIPv6 = ip.To4() == nil
				case *net.IPAddr:
					ip = v.IP
					network = v.String()
					isIPv6 = ip.To4() == nil
				}

				if ip == nil {
					continue
				}

				// Include all interfaces, but mark loopback
				isLoopback := ip.IsLoopback()

				// Get interface flags
				flags := []string{}
				if iface.Flags&net.FlagUp != 0 {
					flags = append(flags, "up")
				}
				if iface.Flags&net.FlagBroadcast != 0 {
					flags = append(flags, "broadcast")
				}
				if iface.Flags&net.FlagLoopback != 0 {
					flags = append(flags, "loopback")
				}
				if iface.Flags&net.FlagPointToPoint != 0 {
					flags = append(flags, "pointtopoint")
				}
				if iface.Flags&net.FlagMulticast != 0 {
					flags = append(flags, "multicast")
				}

				interfaces = append(interfaces, InterfaceInfo{
					Name:     iface.Name,
					IP:       ip.String(),
					Netmask:  network,
					MAC:      iface.HardwareAddr.String(),
					IsIPv6:   isIPv6,
					IsLoopback: isLoopback,
					Flags:    strings.Join(flags, ","),
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

// CollectRoutes collects routing table information
func (c *Collector) CollectRoutes() ([]RouteInfo, error) {
	routes := []RouteInfo{}

	if runtime.GOOS != "linux" {
		return routes, fmt.Errorf("route collection only supported on Linux")
	}

	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return routes, fmt.Errorf("failed to read /proc/net/route: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header
		}

		parts := strings.Fields(line)
		if len(parts) < 11 {
			continue
		}

		// Parse route entry
		// Format: Iface Destination Gateway Flags RefCnt Use Metric Mask MTU Window IRTT
		iface := parts[0]
		destHex := parts[1]
		gatewayHex := parts[2]
		flagsHex := parts[3]
		metricStr := parts[6]

		// Convert hex to IP addresses
		destIP := hexToIP(destHex)
		gatewayIP := hexToIP(gatewayHex)

		// Parse flags
		flags, err := strconv.ParseUint(flagsHex, 16, 32)
		if err != nil {
			continue
		}
		flagsStr := parseRouteFlags(uint32(flags))

		// Parse metric
		metric, err := strconv.Atoi(metricStr)
		if err != nil {
			metric = 0
		}

		// Format destination
		destination := destIP.String()
		if destIP.String() == "0.0.0.0" {
			destination = "default"
		}

		// Format gateway
		gateway := gatewayIP.String()
		if gatewayIP.String() == "0.0.0.0" {
			gateway = "*"
		}

		routes = append(routes, RouteInfo{
			Destination: destination,
			Gateway:     gateway,
			Interface:   iface,
			Flags:       flagsStr,
			Metric:      metric,
		})
	}

	return routes, nil
}

// hexToIP converts a hex string (little-endian) to an IP address
func hexToIP(hexStr string) net.IP {
	val, err := strconv.ParseUint(hexStr, 16, 32)
	if err != nil {
		return net.IPv4(0, 0, 0, 0)
	}

	// Convert from little-endian
	ip := make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, uint32(val))
	return ip
}

// parseRouteFlags converts route flags to human-readable string
func parseRouteFlags(flags uint32) string {
	flagParts := []string{}

	if flags&0x0001 != 0 {
		flagParts = append(flagParts, "U") // Up
	}
	if flags&0x0002 != 0 {
		flagParts = append(flagParts, "G") // Gateway
	}
	if flags&0x0004 != 0 {
		flagParts = append(flagParts, "H") // Host
	}
	if flags&0x0008 != 0 {
		flagParts = append(flagParts, "R") // Reinstate
	}
	if flags&0x0010 != 0 {
		flagParts = append(flagParts, "D") // Dynamic
	}
	if flags&0x0020 != 0 {
		flagParts = append(flagParts, "M") // Modified
	}

	if len(flagParts) == 0 {
		return "none"
	}
	return strings.Join(flagParts, ",")
}

