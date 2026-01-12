package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

type Manager struct {
	dataDir string
	key     *rsa.PrivateKey
	cert    *x509.Certificate
}

func NewManager(dataDir string) (*Manager, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	m := &Manager{
		dataDir: dataDir,
	}

	// Try to load existing key and cert
	if err := m.loadIdentity(); err != nil {
		// If no identity exists, generate new one
		if err := m.generateIdentity(); err != nil {
			return nil, fmt.Errorf("failed to generate identity: %w", err)
		}
	}

	return m, nil
}

func (m *Manager) loadIdentity() error {
	keyPath := filepath.Join(m.dataDir, "agent.key")
	certPath := filepath.Join(m.dataDir, "agent.crt")

	// Load private key
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return fmt.Errorf("failed to decode private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Load certificate
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}

	block, _ = pem.Decode(certData)
	if block == nil {
		return fmt.Errorf("failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	m.key = key
	m.cert = cert
	return nil
}

func (m *Manager) generateIdentity() error {
	// Generate RSA key pair
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Create certificate template (will be signed by control plane)
	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "metanexus-agent",
		},
	}

	// Generate CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, template, key)
	if err != nil {
		return fmt.Errorf("failed to create CSR: %w", err)
	}

	// Save private key
	keyPath := filepath.Join(m.dataDir, "agent.key")
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyFile.Close()

	if err := pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	// Save CSR
	csrPath := filepath.Join(m.dataDir, "agent.csr")
	csrFile, err := os.OpenFile(csrPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create CSR file: %w", err)
	}
	defer csrFile.Close()

	if err := pem.Encode(csrFile, &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	}); err != nil {
		return fmt.Errorf("failed to encode CSR: %w", err)
	}

	m.key = key
	return nil
}

func (m *Manager) GetCSR() ([]byte, error) {
	csrPath := filepath.Join(m.dataDir, "agent.csr")
	return os.ReadFile(csrPath)
}

func (m *Manager) GetPrivateKey() *rsa.PrivateKey {
	return m.key
}

func (m *Manager) GetCertificate() *x509.Certificate {
	return m.cert
}

func (m *Manager) SaveCertificate(certPEM []byte) error {
	certPath := filepath.Join(m.dataDir, "agent.crt")
	return os.WriteFile(certPath, certPEM, 0644)
}

