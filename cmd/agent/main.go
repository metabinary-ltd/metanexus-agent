package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/metabinary-ltd/metanexus-agent/internal/agent"
	"github.com/metabinary-ltd/metanexus-agent/internal/identity"
)

var (
	configPath       = flag.String("config", "/etc/metanexus/agent.yaml", "Path to agent configuration file")
	version          = flag.Bool("version", false, "Print version and exit")
	pairingToken     = flag.String("pairing-token", "", "Pairing token for initial agent registration")
	controlPlaneURL  = flag.String("control-plane-url", "", "Control plane URL for pairing")
)

const (
	Version = "0.1.0"
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("metanexus-agent v%s\n", Version)
		os.Exit(0)
	}

	// Handle pairing mode
	if *pairingToken != "" {
		if *controlPlaneURL == "" {
			log.Fatal("control-plane-url is required when using pairing-token")
		}
		if err := handlePairing(*controlPlaneURL, *pairingToken, *configPath); err != nil {
			log.Fatalf("Pairing failed: %v", err)
		}
		log.Println("Pairing successful! You can now run the agent normally.")
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, gracefully stopping...")
		cancel()
	}()

	// Initialize and run agent
	agt, err := agent.New(*configPath)
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}

	if err := agt.Run(ctx); err != nil {
		log.Fatalf("Agent error: %v", err)
	}
}

func handlePairing(controlPlaneURL, pairingToken, configPath string) error {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	// Determine data directory (use default for pairing)
	dataDir := "/var/lib/metanexus-agent"
	if configPath != "" {
		// TODO: Load from config file if needed
	}

	// Initialize identity manager
	identityMgr, err := identity.NewManager(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize identity: %w", err)
	}

	// Perform pairing
	if err := identityMgr.Pair(controlPlaneURL, pairingToken, hostname); err != nil {
		return fmt.Errorf("pairing failed: %w", err)
	}

	return nil
}
