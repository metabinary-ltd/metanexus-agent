package identity

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type PairingResponse struct {
	AgentID      string `json:"agent_id"`
	Certificate  string `json:"certificate"`
	CACertificate string `json:"ca_certificate"`
}

func (m *Manager) Pair(controlPlaneURL string, pairingToken string, hostname string, agentID *string) error {
	// Get CSR
	csrPEM, err := m.GetCSR()
	if err != nil {
		return fmt.Errorf("failed to get CSR: %w", err)
	}

	// Register with control plane
	reqBody := map[string]string{
		"pairing_token": pairingToken,
		"csr":           string(csrPEM),
		"hostname":      hostname,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/agents/register", controlPlaneURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", string(body))
	}

	var pairingResp PairingResponse
	if err := json.NewDecoder(resp.Body).Decode(&pairingResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Save certificate
	if err := m.SaveCertificate([]byte(pairingResp.Certificate)); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	// Save CA certificate
	caCertPath := filepath.Join(m.dataDir, "ca.crt")
	if err := os.WriteFile(caCertPath, []byte(pairingResp.CACertificate), 0644); err != nil {
		return fmt.Errorf("failed to save CA certificate: %w", err)
	}

	// Save agent ID
	agentIDPath := filepath.Join(m.dataDir, "agent.id")
	if err := os.WriteFile(agentIDPath, []byte(pairingResp.AgentID), 0644); err != nil {
		return fmt.Errorf("failed to save agent ID: %w", err)
	}

	// Reload identity to include certificate
	if err := m.loadIdentity(); err != nil {
		return fmt.Errorf("failed to reload identity: %w", err)
	}

	return nil
}

func (m *Manager) GetTLSConfig() (*tls.Config, error) {
	if m.cert == nil || m.key == nil {
		return nil, fmt.Errorf("certificate or key not loaded")
	}

	// Load CA certificate
	caCertPath := filepath.Join(m.dataDir, "ca.crt")
	caCertData, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	block, _ := pem.Decode(caCertData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA certificate")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	caPool.AddCert(caCert)

	cert := tls.Certificate{
		Certificate: [][]byte{m.cert.Raw},
		PrivateKey:  m.key,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

