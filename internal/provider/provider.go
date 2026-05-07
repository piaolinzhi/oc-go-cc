package provider

import (
	"fmt"
	"sync"

	"oc-go-cc/internal/config"
)

var (
	providers = make(map[string]Provider)
	mu        sync.RWMutex
)

// Provider defines the interface for a service provider.
type Provider interface {
	Name() string
	Config() config.ProviderConfig
	EndpointConfig(modelID string) EndpointConfig
}

// EndpointConfig defines the configuration for an endpoint.
type EndpointConfig struct {
	BaseURL    string
	APIKey     string
	EndpointType string
	TimeoutMs  int
}

// Register registers a provider.
func Register(p Provider) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := providers[p.Name()]; exists {
		return fmt.Errorf("provider %q already registered", p.Name())
	}
	providers[p.Name()] = p
	return nil
}

// Get returns a provider by name.
func Get(name string) (Provider, error) {
	mu.RLock()
	defer mu.RUnlock()

	p, exists := providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %q not found", name)
	}
	return p, nil
}

// List returns all registered providers.
func List() []Provider {
	mu.RLock()
	defer mu.RUnlock()

	list := make([]Provider, 0, len(providers))
	for _, p := range providers {
		list = append(list, p)
	}
	return list
}

// DefaultProvider implements the default OpenCode Go provider.
type DefaultProvider struct {
	name   string
	config config.ProviderConfig
}

// NewDefaultProvider creates a new default provider.
func NewDefaultProvider(name string, cfg config.ProviderConfig) *DefaultProvider {
	return &DefaultProvider{
		name:   name,
		config: cfg,
	}
}

func (p *DefaultProvider) Name() string {
	return p.name
}

func (p *DefaultProvider) Config() config.ProviderConfig {
	return p.config
}

func (p *DefaultProvider) EndpointConfig(modelID string) EndpointConfig {
	cfg := EndpointConfig{
		BaseURL:    p.config.BaseURL,
		APIKey:     p.config.APIKey,
		EndpointType: p.config.EndpointType,
		TimeoutMs:  p.config.TimeoutMs,
	}
	
	if cfg.EndpointType == "" {
		cfg.EndpointType = detectEndpointType(modelID)
	}
	
	if cfg.EndpointType == "anthropic" && p.config.AnthropicBaseURL != "" {
		cfg.BaseURL = p.config.AnthropicBaseURL
	}
	
	return cfg
}

// detectEndpointType determines the endpoint type based on model ID.
func detectEndpointType(modelID string) string {
	switch modelID {
	case "minimax-m2.5", "minimax-m2.7":
		return "anthropic"
	default:
		return "openai"
	}
}