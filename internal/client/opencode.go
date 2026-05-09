// Package client manages upstream API client connections.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"oc-go-cc/internal/config"
	"oc-go-cc/internal/provider"
	"oc-go-cc/pkg/types"
)

// OpenCodeClient handles communication with OpenCode Go API.
type OpenCodeClient struct {
	defaultProvider provider.Provider
	atomic           *config.AtomicConfig
	httpClient       *http.Client
	logger           *slog.Logger
}

// EndpointConfig holds configuration for a specific API endpoint.
type EndpointConfig struct {
	BaseURL      string
	APIKey       string
	EndpointType string
}

// NewOpenCodeClient creates a new OpenCode Go client.
func NewOpenCodeClient(atomic *config.AtomicConfig) *OpenCodeClient {
	cfg := atomic.Get()
	timeout := time.Duration(cfg.OpenCodeGo.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	return &OpenCodeClient{
		atomic:    atomic,
		httpClient: createHTTPClient(timeout),
		logger:    slog.Default(),
	}
}

// NewOpenCodeClientWithProvider creates a new OpenCode Go client with a specific provider.
func NewOpenCodeClientWithProvider(p provider.Provider) *OpenCodeClient {
	timeout := time.Duration(p.Config().TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	return &OpenCodeClient{
		defaultProvider: p,
		httpClient:      createHTTPClient(timeout),
		logger:          slog.Default(),
	}
}

func createHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		MaxConnsPerHost:     50,
		DisableKeepAlives:   false,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// IsAnthropicModel returns true if the model requires the Anthropic endpoint.
func IsAnthropicModel(modelID string) bool {
	switch modelID {
	case "minimax-m2.5", "minimax-m2.7":
		return true
	default:
		return false
	}
}

// getEndpoint returns the appropriate endpoint config for a model and provider.
func (c *OpenCodeClient) getEndpoint(modelID string, providerName string) EndpointConfig {
	var p provider.Provider
	var err error

	c.logger.Debug("getEndpoint called", "modelID 【", modelID, "】 providerName【", providerName,"】")

	// Try to get provider by name if specified
	if providerName != "" {
		p, err = provider.Get(providerName)
		if err == nil {
			ep := p.EndpointConfig(modelID)
			c.logger.Debug("using provider", "provider", providerName, "baseURL", ep.BaseURL, "endpointType", ep.EndpointType)
			return EndpointConfig{
				BaseURL:      ep.BaseURL,
				APIKey:       ep.APIKey,
				EndpointType: ep.EndpointType,
			}
		}
		c.logger.Error("provider not found", "provider", providerName, "error", err)
	}

	// Fall back to default provider
	if c.defaultProvider != nil {
		ep := c.defaultProvider.EndpointConfig(modelID)
		c.logger.Debug("Using default provider", "baseURL", ep.BaseURL, "endpointType", ep.EndpointType)
		return EndpointConfig{
			BaseURL:      ep.BaseURL,
			APIKey:       ep.APIKey,
			EndpointType: ep.EndpointType,
		}
	}

	// Default behavior based on model type
	if IsAnthropicModel(modelID) {
		return EndpointConfig{EndpointType: "anthropic"}
	}
	return EndpointConfig{EndpointType: "openai"}
}

// ChatCompletion sends a chat completion request to the OpenCode Go API.
// Returns the raw HTTP response for the caller to handle (streaming or body read).
func (c *OpenCodeClient) ChatCompletion(
	ctx context.Context,
	modelID string,
	req *types.ChatCompletionRequest,
	apiKey string,
	providerName string,
) (*http.Response, error) {
	endpoint := c.getEndpoint(modelID, providerName)
	
	if endpoint.BaseURL == "" {
		return nil, fmt.Errorf("no base URL configured for model %s", modelID)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Use provider API key if available, otherwise fallback to global API key
	key := endpoint.APIKey
	if key == "" {
		key = apiKey
	}
	httpReq.Header.Set("Authorization", "Bearer "+key)

	// Add streaming header if requested
	if req.Stream != nil && *req.Stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode >= http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// ChatCompletionNonStreaming sends a non-streaming request and returns the full parsed response.
func (c *OpenCodeClient) ChatCompletionNonStreaming(
	ctx context.Context,
	modelID string,
	req *types.ChatCompletionRequest,
	apiKey string,
	providerName string,
) (*types.ChatCompletionResponse, error) {
	// Force non-streaming
	streamFalse := false
	req.Stream = &streamFalse

	resp, err := c.ChatCompletion(ctx, modelID, req, apiKey, providerName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp types.ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// GetStreamingBody returns the response body for streaming consumption.
// The caller is responsible for closing the returned ReadCloser.
func (c *OpenCodeClient) GetStreamingBody(
	ctx context.Context,
	modelID string,
	req *types.ChatCompletionRequest,
	apiKey string,
	providerName string,
) (io.ReadCloser, error) {
	// Force streaming
	streamTrue := true
	req.Stream = &streamTrue

	resp, err := c.ChatCompletion(ctx, modelID, req, apiKey, providerName)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// SendAnthropicRequest sends a raw Anthropic-format request (for MiniMax models).
// This skips the OpenAI transformation entirely.
func (c *OpenCodeClient) SendAnthropicRequest(
	ctx context.Context,
	body []byte,
	stream bool,
	apiKey string,
	providerName string,
) (*http.Response, error) {
	// Extract model from body to get correct endpoint config
	var req struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request body for model extraction: %w", err)
	}
	c.logger.Info("SendAnthropicRequest", "model【", req.Model, "】provider", providerName)
	
	endpoint := c.getEndpoint(req.Model, providerName)
	
	if endpoint.BaseURL == "" {
		return nil, fmt.Errorf("no base URL configured for anthropic endpoint")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Use provider API key if available, otherwise fallback to global API key
	key := endpoint.APIKey
	if key == "" {
		key = apiKey
	}
	httpReq.Header.Set("Authorization", "Bearer "+key)
	httpReq.Header.Set("x-api-key", key)

	// Add streaming header if requested
	if stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode >= http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}
