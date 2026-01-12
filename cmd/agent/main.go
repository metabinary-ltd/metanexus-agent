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
)

var (
	configPath = flag.String("config", "/etc/metanexus/agent.yaml", "Path to agent configuration file")
	version    = flag.Bool("version", false, "Print version and exit")
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
