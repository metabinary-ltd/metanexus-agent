package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

func (m *Manager) GetAgentID() (string, error) {
	// Try to read stored agent ID
	agentIDPath := filepath.Join(m.dataDir, "agent.id")
	if data, err := os.ReadFile(agentIDPath); err == nil {
		return string(data), nil
	}

	// Generate agent ID from certificate fingerprint if available
	if m.cert != nil {
		fingerprint := sha256.Sum256(m.cert.Raw)
		agentID := hex.EncodeToString(fingerprint[:])
		
		// Store it
		os.WriteFile(agentIDPath, []byte(agentID), 0644)
		return agentID, nil
	}

	return "", fmt.Errorf("no agent ID available - agent not paired")
}

