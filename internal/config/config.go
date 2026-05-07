// Package config handles application configuration loading and validation.
package config

import "encoding/json"

// Config holds the complete application configuration.
type Config struct {
	APIKey         string                   `json:"api_key"`
	Host           string                   `json:"host"`
	Port           int                      `json:"port"`
	Models         map[string]ModelConfig   `json:"models"`
	Fallbacks      map[string][]ModelConfig `json:"fallbacks"`
	Providers      map[string]ProviderConfig `json:"providers,omitempty"`
	OpenCodeGo     OpenCodeGoConfig         `json:"opencode_go"`
	Logging        LoggingConfig            `json:"logging"`
}

// ModelConfig defines routing rules for a specific model.
type ModelConfig struct {
	Provider         string          `json:"provider"`
	ModelID          string          `json:"model_id"`
	Temperature      float64         `json:"temperature"`
	MaxTokens        int             `json:"max_tokens"`
	ContextThreshold int             `json:"context_threshold"`
	ReasoningEffort  string          `json:"reasoning_effort"`
	Thinking         json.RawMessage `json:"thinking,omitempty"`
}

// ProviderConfig defines configuration for a single provider.
type ProviderConfig struct {
	Name             string `json:"name"`
	BaseURL          string `json:"base_url"`
	AnthropicBaseURL string `json:"anthropic_base_url,omitempty"`
	APIKey           string `json:"api_key,omitempty"`
	EndpointType     string `json:"endpoint_type,omitempty"`
	TimeoutMs        int    `json:"timeout_ms,omitempty"`
}

// OpenCodeGoConfig holds the upstream OpenCode Go API settings.
type OpenCodeGoConfig struct {
	BaseURL          string `json:"base_url"`
	AnthropicBaseURL string `json:"anthropic_base_url"`
	TimeoutMs        int    `json:"timeout_ms"`
}

// LoggingConfig controls application logging behavior.
type LoggingConfig struct {
	Level    string `json:"level"`
	Requests bool   `json:"requests"`
}
