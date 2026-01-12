package transport

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/metanexus/metanexus-agent/internal/identity"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	identity   *identity.Manager
}

func NewClient(baseURL string, identityMgr *identity.Manager) (*Client, error) {
	client := &Client{
		baseURL:  baseURL,
		identity: identityMgr,
	}

	// Try to get TLS config (may fail if not paired)
	tlsConfig, err := identityMgr.GetTLSConfig()
	if err != nil {
		// Not paired yet, use insecure connection
		client.httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
		return client, nil
	}

	// Use mTLS
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client.httpClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return client, nil
}

func (c *Client) Get(url string) (*http.Response, error) {
	return c.httpClient.Get(c.baseURL + url)
}

func (c *Client) Post(url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", c.baseURL+url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *Client) UpdateTLSConfig() error {
	tlsConfig, err := c.identity.GetTLSConfig()
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	c.httpClient.Transport = transport
	return nil
}
