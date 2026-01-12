package actions

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
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

type ExecuteCommandVerb struct {
	// No dependencies needed
}

func NewExecuteCommandVerb() *ExecuteCommandVerb {
	return &ExecuteCommandVerb{}
}

func extractString(params map[string]interface{}, key string, defaultValue string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return defaultValue
}

func extractInt(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	if val, ok := params[key].(int); ok {
		return val
	}
	return defaultValue
}

func extractBool(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key].(bool); ok {
		return val
	}
	return defaultValue
}

func extractStringArray(params map[string]interface{}, key string) []string {
	if val, ok := params[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}

func extractMap(params map[string]interface{}, key string) map[string]string {
	if val, ok := params[key].(map[string]interface{}); ok {
		result := make(map[string]string)
		for k, v := range val {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result
	}
	return nil
}

func (v *ExecuteCommandVerb) Execute(params map[string]interface{}) (map[string]interface{}, error) {
	// Extract parameters
	command := extractString(params, "command", "")
	if command == "" {
		return nil, fmt.Errorf("command parameter is required")
	}

	args := extractStringArray(params, "args")
	workingDir := extractString(params, "working_directory", "")
	timeout := extractInt(params, "timeout_seconds", 300)
	env := extractMap(params, "environment")
	stdin := extractString(params, "stdin", "")
	userStr := extractString(params, "user", "")
	groupStr := extractString(params, "group", "")
	captureStdout := extractBool(params, "capture_stdout", true)
	captureStderr := extractBool(params, "capture_stderr", true)
	expectedExitCode := extractInt(params, "expected_exit_code", 0)
	retryCount := extractInt(params, "retry_count", 0)
	retryDelay := extractInt(params, "retry_delay_seconds", 1)

	// Execute with retry logic
	var lastErr error
	var lastResult map[string]interface{}

	for attempt := 0; attempt <= retryCount; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}

		result, err := v.executeCommand(
			command, args, workingDir, timeout, env, stdin,
			userStr, groupStr, captureStdout, captureStderr,
		)

		if err != nil {
			lastErr = err
			lastResult = result
			continue
		}

		// Check exit code if specified
		if exitCode, ok := result["exit_code"].(int); ok {
			if exitCode != expectedExitCode {
				lastErr = fmt.Errorf("command exited with code %d, expected %d", exitCode, expectedExitCode)
				lastResult = result
				continue
			}
		}

		return result, nil
	}

	// All retries failed
	if lastResult != nil {
		return lastResult, lastErr
	}
	return nil, lastErr
}

func (v *ExecuteCommandVerb) executeCommand(
	command string,
	args []string,
	workingDir string,
	timeoutSeconds int,
	env map[string]string,
	stdin string,
	userStr string,
	groupStr string,
	captureStdout bool,
	captureStderr bool,
) (map[string]interface{}, error) {
	startTime := time.Now()

	// Create command context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, command, args...)

	// Set working directory
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Set environment variables
	if len(env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Set user/group if specified
	if userStr != "" || groupStr != "" {
		var uid, gid int

		if userStr != "" {
			u, err := user.Lookup(userStr)
			if err != nil {
				return nil, fmt.Errorf("failed to lookup user %s: %w", userStr, err)
			}
			uid, err = strconv.Atoi(u.Uid)
			if err != nil {
				return nil, fmt.Errorf("failed to parse user UID: %w", err)
			}
		}

		if groupStr != "" {
			g, err := user.LookupGroup(groupStr)
			if err != nil {
				return nil, fmt.Errorf("failed to lookup group %s: %w", groupStr, err)
			}
			gid, err = strconv.Atoi(g.Gid)
			if err != nil {
				return nil, fmt.Errorf("failed to parse group GID: %w", err)
			}
		}

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(uid),
				Gid: uint32(gid),
			},
		}
	}

	// Set up stdin
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	// Capture output
	var stdout, stderr strings.Builder
	if captureStdout {
		cmd.Stdout = &stdout
	} else {
		cmd.Stdout = nil
	}

	if captureStderr {
		cmd.Stderr = &stderr
	} else {
		cmd.Stderr = nil
	}

	// Execute command
	err := cmd.Run()
	duration := time.Since(startTime)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// Context timeout or other error
			if ctx.Err() == context.DeadlineExceeded {
				return map[string]interface{}{
					"status":    "timeout",
					"error":      "command execution timed out",
					"duration":   duration.Seconds(),
					"exit_code":  -1,
					"stdout":     stdout.String(),
					"stderr":     stderr.String(),
				}, fmt.Errorf("command execution timed out after %d seconds", timeoutSeconds)
			}
			return map[string]interface{}{
				"status":    "error",
				"error":     err.Error(),
				"duration":  duration.Seconds(),
				"exit_code": -1,
				"stdout":    stdout.String(),
				"stderr":    stderr.String(),
			}, err
		}
	}

	result := map[string]interface{}{
		"status":    "completed",
		"exit_code": exitCode,
		"duration":  duration.Seconds(),
	}

	if captureStdout {
		result["stdout"] = stdout.String()
	}
	if captureStderr {
		result["stderr"] = stderr.String()
	}

	return result, nil
}

type ListNetworkInterfacesVerb struct {
	collector *collect.Collector
}

func NewListNetworkInterfacesVerb() *ListNetworkInterfacesVerb {
	return &ListNetworkInterfacesVerb{
		collector: collect.New(),
	}
}

func (v *ListNetworkInterfacesVerb) Execute(params map[string]interface{}) (map[string]interface{}, error) {
	networkData, err := v.collector.CollectNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to collect network data: %w", err)
	}

	// Format interfaces for response
	interfaces := make([]map[string]interface{}, len(networkData.Interfaces))
	for i, iface := range networkData.Interfaces {
		interfaces[i] = map[string]interface{}{
			"name":        iface.Name,
			"ip":          iface.IP,
			"network":     iface.Netmask,
			"mac":         iface.MAC,
			"is_ipv6":     iface.IsIPv6,
			"is_loopback": iface.IsLoopback,
			"flags":       iface.Flags,
		}
	}

	return map[string]interface{}{
		"status":     "completed",
		"interfaces": interfaces,
		"count":      len(interfaces),
	}, nil
}

type ListARPNeighborsVerb struct {
	collector *collect.Collector
}

func NewListARPNeighborsVerb() *ListARPNeighborsVerb {
	return &ListARPNeighborsVerb{
		collector: collect.New(),
	}
}

func (v *ListARPNeighborsVerb) Execute(params map[string]interface{}) (map[string]interface{}, error) {
	networkData, err := v.collector.CollectNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to collect network data: %w", err)
	}

	// Format ARP neighbors for response
	neighbors := make([]map[string]interface{}, len(networkData.ARPNeighbors))
	for i, neighbor := range networkData.ARPNeighbors {
		neighbors[i] = map[string]interface{}{
			"ip":        neighbor.IP,
			"mac":       neighbor.MAC,
			"interface": neighbor.Interface,
		}
	}

	return map[string]interface{}{
		"status":    "completed",
		"neighbors": neighbors,
		"count":     len(neighbors),
	}, nil
}

type ListRoutesVerb struct {
	collector *collect.Collector
}

func NewListRoutesVerb() *ListRoutesVerb {
	return &ListRoutesVerb{
		collector: collect.New(),
	}
}

func (v *ListRoutesVerb) Execute(params map[string]interface{}) (map[string]interface{}, error) {
	routes, err := v.collector.CollectRoutes()
	if err != nil {
		return nil, fmt.Errorf("failed to collect routes: %w", err)
	}

	// Format routes for response
	routeList := make([]map[string]interface{}, len(routes))
	for i, route := range routes {
		routeList[i] = map[string]interface{}{
			"destination": route.Destination,
			"gateway":     route.Gateway,
			"interface":   route.Interface,
			"flags":       route.Flags,
			"metric":      route.Metric,
		}
	}

	return map[string]interface{}{
		"status": "completed",
		"routes": routeList,
		"count":  len(routeList),
	}, nil
}

type CheckUpdatesVerb struct {
	collector *collect.Collector
}

func NewCheckUpdatesVerb() *CheckUpdatesVerb {
	return &CheckUpdatesVerb{
		collector: collect.New(),
	}
}

func (v *CheckUpdatesVerb) Execute(params map[string]interface{}) (map[string]interface{}, error) {
	updateInfo, err := v.collector.CollectUpdates()
	if err != nil {
		// Return partial result if package manager not found
		if updateInfo != nil && updateInfo.PackageManager == "none" {
			return map[string]interface{}{
				"status":          "completed",
				"package_manager": "none",
				"update_count":    0,
				"packages":         []interface{}{},
				"message":          "No supported package manager found",
			}, nil
		}
		return nil, fmt.Errorf("failed to check updates: %w", err)
	}

	// Format packages for response
	packages := make([]map[string]interface{}, len(updateInfo.Packages))
	for i, pkg := range updateInfo.Packages {
		packages[i] = map[string]interface{}{
			"name":      pkg.Name,
			"current":   pkg.Current,
			"available": pkg.Available,
		}
	}

	return map[string]interface{}{
		"status":          "completed",
		"package_manager": updateInfo.PackageManager,
		"update_count":    updateInfo.UpdateCount,
		"packages":        packages,
		"last_checked":    updateInfo.LastChecked.Format(time.RFC3339),
	}, nil
}
