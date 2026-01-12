package actions

import (
	"fmt"

	"github.com/metanexus/metanexus-agent/internal/identity"
	"github.com/metanexus/metanexus-agent/internal/transport"
)

type Executor struct {
	identity  *identity.Manager
	transport *transport.Client
	verbs     map[string]VerbHandler
}

func NewExecutor(identityMgr *identity.Manager, transportClient *transport.Client) *Executor {
	exec := &Executor{
		identity:  identityMgr,
		transport: transportClient,
		verbs:     make(map[string]VerbHandler),
	}

	// Register known verbs
	updateContainerVerb, _ := NewUpdateContainerVerb()
	if updateContainerVerb != nil {
		exec.RegisterVerb("update-container", updateContainerVerb)
	}
	exec.RegisterVerb("health-snapshot", NewHealthSnapshotVerb())
	exec.RegisterVerb("rotate-identity", NewRotateIdentityVerb(identityMgr))

	return exec
}

type VerbHandler interface {
	Execute(params map[string]interface{}) (map[string]interface{}, error)
}

func (e *Executor) RegisterVerb(name string, handler VerbHandler) {
	e.verbs[name] = handler
}

func (e *Executor) Execute(actionID string, verb string, params map[string]interface{}) (map[string]interface{}, error) {
	handler, ok := e.verbs[verb]
	if !ok {
		return nil, fmt.Errorf("unknown verb: %s", verb)
	}

	return handler.Execute(params)
}

// Verb implementations are in verbs.go

