package actions

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/metabinary-ltd/metanexus-agent/internal/transport"
)

type ActionCatalog struct {
	Version   string           `json:"version"`
	IssuedAt  string           `json:"issued_at"`
	Actions   []ActionDef      `json:"actions"`
	Signature CatalogSignature `json:"signature"`
}

type ActionDef struct {
	ID                   string                 `json:"id"`
	Verb                 string                 `json:"verb"`
	Name                 string                 `json:"name"`
	Description          string                 `json:"description"`
	YAML                 string                 `json:"yaml"`
	RequiredCapabilities []string               `json:"required_capabilities,omitempty"`
	Parameters           map[string]interface{} `json:"parameters,omitempty"`
}

type CatalogSignature struct {
	Algorithm   string `json:"algorithm"`
	Value       string `json:"value"`
	Certificate string `json:"certificate,omitempty"`
}

func FetchCatalog(client *transport.Client) (*ActionCatalog, error) {
	resp, err := client.Get("/api/v1/action-catalog")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog fetch failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var catalog ActionCatalog
	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog: %w", err)
	}

	return &catalog, nil
}

func ValidateCatalog(catalog *ActionCatalog, caCertPEM []byte) error {
	// Parse CA certificate
	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		return fmt.Errorf("failed to decode CA certificate")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Parse signature certificate if present
	var sigCert *x509.Certificate
	if catalog.Signature.Certificate != "" {
		block, _ := pem.Decode([]byte(catalog.Signature.Certificate))
		if block != nil {
			sigCert, err = x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse signature certificate: %w", err)
			}

			// Verify signature certificate is signed by CA
			if err := sigCert.CheckSignatureFrom(caCert); err != nil {
				return fmt.Errorf("signature certificate not signed by CA: %w", err)
			}
		}
	}

	// Reconstruct payload for verification (without signature)
	payload := map[string]interface{}{
		"version":   catalog.Version,
		"issued_at": catalog.IssuedAt,
		"actions":   catalog.Actions,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Verify signature
	signature, err := base64.StdEncoding.DecodeString(catalog.Signature.Value)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	hash := sha256.Sum256(payloadJSON)

	// Use CA cert's public key if no signature cert
	certToUse := caCert
	if sigCert != nil {
		certToUse = sigCert
	}

	pubKey, ok := certToUse.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("unsupported public key type")
	}

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// Check expiration if present
	if catalog.IssuedAt != "" {
		issuedAt, err := time.Parse(time.RFC3339, catalog.IssuedAt)
		if err == nil {
			// Catalog should be recent (within 24 hours for MVP)
			if time.Since(issuedAt) > 24*time.Hour {
				return fmt.Errorf("catalog is too old")
			}
		}
	}

	return nil
}
