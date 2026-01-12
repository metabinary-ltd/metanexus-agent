package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AgentID          string        `yaml:"agent_id"`
	ControlPlaneURL  string        `yaml:"control_plane_url"`
	DataDir          string        `yaml:"data_dir"`
	TelemetryInterval time.Duration `yaml:"telemetry_interval"`
}

func Load(configPath string) (*Config, error) {
	// Default config
	config := &Config{
		ControlPlaneURL:   "http://localhost:3000",
		DataDir:           "/var/lib/metanexus-agent",
		TelemetryInterval: 30 * time.Second,
	}

	// Try to load from file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return default config
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func Save(configPath string, config *Config) error {
	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := ""
	if len(configPath) > 0 {
		dir = configPath[:len(configPath)-len(getFileName(configPath))]
	}
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func getFileName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

