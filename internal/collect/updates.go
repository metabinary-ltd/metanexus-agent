package collect

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type UpdateInfo struct {
	PackageManager string
	UpdateCount    int
	Packages       []PackageUpdate
	LastChecked    time.Time
}

type PackageUpdate struct {
	Name      string
	Current   string
	Available string
}

// CollectUpdates checks for available system updates
func (c *Collector) CollectUpdates() (*UpdateInfo, error) {
	info := &UpdateInfo{
		LastChecked: time.Now(),
	}

	// Detect package manager
	if runtime.GOOS != "linux" {
		info.PackageManager = "none"
		return info, fmt.Errorf("update checking only supported on Linux")
	}

	// Try apt first (Debian/Ubuntu)
	if _, err := exec.LookPath("apt"); err == nil {
		return c.checkAptUpdates(info)
	}

	// Try dnf (Fedora/RHEL 8+)
	if _, err := exec.LookPath("dnf"); err == nil {
		return c.checkDnfUpdates(info)
	}

	// Try yum (RHEL/CentOS 7)
	if _, err := exec.LookPath("yum"); err == nil {
		return c.checkYumUpdates(info)
	}

	info.PackageManager = "none"
	return info, fmt.Errorf("no supported package manager found (apt, dnf, or yum)")
}

func (c *Collector) checkAptUpdates(info *UpdateInfo) (*UpdateInfo, error) {
	info.PackageManager = "apt"

	// Run apt list --upgradable (non-interactive)
	cmd := exec.Command("apt", "list", "--upgradable")
	output, err := cmd.Output()
	if err != nil {
		// apt list returns exit code 0 even when there are no updates
		// Check if output is empty or just headers
		outputStr := string(output)
		if strings.TrimSpace(outputStr) == "" || !strings.Contains(outputStr, "/") {
			info.UpdateCount = 0
			info.Packages = []PackageUpdate{}
			return info, nil
		}
	}

	// Parse output
	// Format: package/version,now/current,upgradable/available
	lines := strings.Split(string(output), "\n")
	packages := []PackageUpdate{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Listing...") {
			continue
		}

		// Parse line: package/version,now/current,upgradable/available
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		// First part: package/version
		pkgParts := strings.Split(parts[0], "/")
		if len(pkgParts) < 2 {
			continue
		}
		pkgName := pkgParts[0]

		// Second part: now/current,upgradable/available
		versionParts := strings.Split(parts[1], ",")
		var current, available string

		for _, vp := range versionParts {
			if strings.Contains(vp, "now/") {
				current = strings.TrimPrefix(vp, "now/")
			} else if strings.Contains(vp, "upgradable/") {
				available = strings.TrimPrefix(vp, "upgradable/")
			}
		}

		if available != "" && available != current {
			packages = append(packages, PackageUpdate{
				Name:      pkgName,
				Current:   current,
				Available: available,
			})
		}
	}

	info.UpdateCount = len(packages)
	info.Packages = packages
	return info, nil
}

func (c *Collector) checkDnfUpdates(info *UpdateInfo) (*UpdateInfo, error) {
	info.PackageManager = "dnf"

	// Run dnf check-update (quiet, non-interactive)
	cmd := exec.Command("dnf", "check-update", "-q")
	output, err := cmd.Output()

	// dnf check-update returns exit code 100 when updates are available
	// exit code 0 means no updates
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok || exitError.ExitCode() != 100 {
			// Real error or no updates
			if exitError != nil && exitError.ExitCode() == 0 {
				info.UpdateCount = 0
				info.Packages = []PackageUpdate{}
				return info, nil
			}
			return info, fmt.Errorf("failed to check dnf updates: %w", err)
		}
	}

	// Parse output
	// Format: package.arch version-release repo
	lines := strings.Split(string(output), "\n")
	packages := []PackageUpdate{}
	seenPackages := make(map[string]bool)

	// Regex to match package lines
	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Last metadata") || strings.HasPrefix(line, "Updating") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			pkgFull := matches[1]
			available := matches[2]

			// Extract package name (remove .arch)
			pkgParts := strings.Split(pkgFull, ".")
			pkgName := pkgParts[0]

			// Skip if we've already seen this package
			if seenPackages[pkgName] {
				continue
			}
			seenPackages[pkgName] = true

			// Get current version (would need rpm -q, but that's expensive)
			// For now, just show available version
			packages = append(packages, PackageUpdate{
				Name:      pkgName,
				Current:   "unknown", // Would require additional query
				Available: available,
			})
		}
	}

	info.UpdateCount = len(packages)
	info.Packages = packages
	return info, nil
}

func (c *Collector) checkYumUpdates(info *UpdateInfo) (*UpdateInfo, error) {
	info.PackageManager = "yum"

	// Run yum check-update (quiet, non-interactive)
	cmd := exec.Command("yum", "check-update", "-q")
	output, err := cmd.Output()

	// yum check-update returns exit code 100 when updates are available
	// exit code 0 means no updates
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok || exitError.ExitCode() != 100 {
			// Real error or no updates
			if exitError != nil && exitError.ExitCode() == 0 {
				info.UpdateCount = 0
				info.Packages = []PackageUpdate{}
				return info, nil
			}
			return info, fmt.Errorf("failed to check yum updates: %w", err)
		}
	}

	// Parse output (similar to dnf)
	lines := strings.Split(string(output), "\n")
	packages := []PackageUpdate{}
	seenPackages := make(map[string]bool)

	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Loaded plugins") || strings.HasPrefix(line, "Updating") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			pkgFull := matches[1]
			available := matches[2]

			pkgParts := strings.Split(pkgFull, ".")
			pkgName := pkgParts[0]

			if seenPackages[pkgName] {
				continue
			}
			seenPackages[pkgName] = true

			packages = append(packages, PackageUpdate{
				Name:      pkgName,
				Current:   "unknown",
				Available: available,
			})
		}
	}

	info.UpdateCount = len(packages)
	info.Packages = packages
	return info, nil
}

