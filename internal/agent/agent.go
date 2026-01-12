package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/metabinary-ltd/metanexus-agent/internal/actions"
	"github.com/metabinary-ltd/metanexus-agent/internal/collect"
	"github.com/metabinary-ltd/metanexus-agent/internal/config"
	"github.com/metabinary-ltd/metanexus-agent/internal/identity"
	"github.com/metabinary-ltd/metanexus-agent/internal/transport"
)

type Config struct {
	ControlPlaneURL   string        `yaml:"control_plane_url"`
	DataDir           string        `yaml:"data_dir"`
	TelemetryInterval time.Duration `yaml:"telemetry_interval"`
}

type Agent struct {
	config     *Config
	agentConfig *config.Config
	identity   *identity.Manager
	transport  *transport.Client
	collector  *collect.Collector
	actionExec *actions.Executor
}

func New(configPath string) (*Agent, error) {
	// Load config from file
	agentConfig, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Convert to internal config format
	internalConfig := &Config{
		ControlPlaneURL:   agentConfig.ControlPlaneURL,
		DataDir:           agentConfig.DataDir,
		TelemetryInterval: agentConfig.TelemetryInterval,
	}

	// Use defaults if not set
	if internalConfig.ControlPlaneURL == "" {
		internalConfig.ControlPlaneURL = "http://localhost:3000"
	}
	if internalConfig.DataDir == "" {
		internalConfig.DataDir = "/var/lib/metanexus-agent"
	}
	if internalConfig.TelemetryInterval == 0 {
		internalConfig.TelemetryInterval = 30 * time.Second
	}

	// Initialize identity manager
	identityMgr, err := identity.NewManager(internalConfig.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize identity: %w", err)
	}

	// Initialize transport client
	transportClient, err := transport.NewClient(internalConfig.ControlPlaneURL, identityMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize transport: %w", err)
	}

	// Initialize collector
	collector := collect.New()

	// Initialize action executor
	actionExec := actions.NewExecutor(identityMgr, transportClient, internalConfig.DataDir)

	return &Agent{
		config:      internalConfig,
		agentConfig: agentConfig,
		identity:    identityMgr,
		transport:   transportClient,
		collector:   collector,
		actionExec:  actionExec,
	}, nil
}

func (a *Agent) Run(ctx context.Context) error {
	log.Println("Starting MetaNexus agent...")

	// Start telemetry reporting loop
	go a.telemetryLoop(ctx)

	// Start action polling loop
	go a.actionLoop(ctx)

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Agent shutting down...")

	return nil
}

func (a *Agent) telemetryLoop(ctx context.Context) {
	ticker := time.NewTicker(a.config.TelemetryInterval)
	defer ticker.Stop()

	// Send initial telemetry
	if err := a.sendTelemetry(ctx); err != nil {
		log.Printf("Failed to send initial telemetry: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.sendTelemetry(ctx); err != nil {
				log.Printf("Failed to send telemetry: %v", err)
			}
		}
	}
}

func (a *Agent) actionLoop(ctx context.Context) {
	// Refresh catalog every 5 minutes
	catalogTicker := time.NewTicker(5 * time.Minute)
	defer catalogTicker.Stop()

	// Poll for actions every 10 seconds
	actionTicker := time.NewTicker(10 * time.Second)
	defer actionTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-catalogTicker.C:
			// Refresh catalog periodically
			if err := a.actionExec.RefreshCatalog(); err != nil {
				log.Printf("Failed to refresh action catalog: %v", err)
			}
		case <-actionTicker.C:
			// TODO: Poll for actions and execute
			_ = a.actionExec
		}
	}
}
