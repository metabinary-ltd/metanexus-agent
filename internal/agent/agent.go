package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/metabinary-ltd/metanexus-agent/internal/actions"
	"github.com/metabinary-ltd/metanexus-agent/internal/collect"
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
	identity   *identity.Manager
	transport  *transport.Client
	collector  *collect.Collector
	actionExec *actions.Executor
}

func New(configPath string) (*Agent, error) {
	// TODO: Load config from file
	config := &Config{
		ControlPlaneURL:   "http://localhost:3000",
		DataDir:           "/var/lib/metanexus-agent",
		TelemetryInterval: 30 * time.Second,
	}

	// Initialize identity manager
	identityMgr, err := identity.NewManager(config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize identity: %w", err)
	}

	// Initialize transport client
	transportClient, err := transport.NewClient(config.ControlPlaneURL, identityMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize transport: %w", err)
	}

	// Initialize collector
	collector := collect.New()

	// Initialize action executor
	actionExec := actions.NewExecutor(identityMgr, transportClient)

	return &Agent{
		config:     config,
		identity:   identityMgr,
		transport:  transportClient,
		collector:  collector,
		actionExec: actionExec,
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
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// TODO: Poll for actions and execute
			_ = a.actionExec
		}
	}
}
