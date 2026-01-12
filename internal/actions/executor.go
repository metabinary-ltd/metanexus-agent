package actions

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/metabinary-ltd/metanexus-agent/internal/identity"
	"github.com/metabinary-ltd/metanexus-agent/internal/transport"
)

type Executor struct {
	identity  *identity.Manager
	transport *transport.Client
	verbs     map[string]VerbHandler
	catalog   *ActionCatalog
	dataDir   string
}

func NewExecutor(identityMgr *identity.Manager, transportClient *transport.Client, dataDir string) *Executor {
	exec := &Executor{
		identity:  identityMgr,
		transport: transportClient,
		verbs:     make(map[string]VerbHandler),
		dataDir:   dataDir,
	}

	// Initialize verb handlers (these are the implementations)
	updateContainerVerb, _ := NewUpdateContainerVerb()
	if updateContainerVerb != nil {
		exec.RegisterVerb("update-container", updateContainerVerb)
	}
	exec.RegisterVerb("health-snapshot", NewHealthSnapshotVerb())
	exec.RegisterVerb("rotate-identity", NewRotateIdentityVerb(identityMgr))

	// Load catalog on startup
	if err := exec.loadCatalog(); err != nil {
		log.Printf("Warning: Failed to load action catalog: %v", err)
		log.Println("Agent will attempt to fetch catalog from control plane")
	}

	return exec
}

type VerbHandler interface {
	Execute(params map[string]interface{}) (map[string]interface{}, error)
}

func (e *Executor) RegisterVerb(name string, handler VerbHandler) {
	e.verbs[name] = handler
}

func (e *Executor) Execute(actionID string, verb string, params map[string]interface{}) (map[string]interface{}, error) {
	// Verify action is in catalog
	if e.catalog != nil {
		found := false
		for _, action := range e.catalog.Actions {
			if action.Verb == verb {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("action verb '%s' not found in catalog", verb)
		}
	}

	handler, ok := e.verbs[verb]
	if !ok {
		return nil, fmt.Errorf("unknown verb: %s", verb)
	}

	return handler.Execute(params)
}

func (e *Executor) loadCatalog() error {
	// Try to fetch catalog from control plane
	catalog, err := FetchCatalog(e.transport)
	if err != nil {
		return fmt.Errorf("failed to fetch catalog: %w", err)
	}

	// Load CA certificate for validation
	caCertPath := filepath.Join(e.dataDir, "ca.crt")
	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// Validate catalog signature
	if err := ValidateCatalog(catalog, caCertPEM); err != nil {
		return fmt.Errorf("catalog validation failed: %w", err)
	}

	e.catalog = catalog
	log.Printf("Loaded action catalog version %s with %d actions", catalog.Version, len(catalog.Actions))
	return nil
}

func (e *Executor) RefreshCatalog() error {
	return e.loadCatalog()
}

func (e *Executor) GetCatalog() *ActionCatalog {
	return e.catalog
}

// Verb implementations are in verbs.go
