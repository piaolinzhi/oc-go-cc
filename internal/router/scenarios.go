package router

import (
	"fmt"
	"strings"

	"oc-go-cc/internal/config"
)

// Scenario represents the routing scenario for model selection.
type Scenario string

const (
	ScenarioDefault     Scenario = "default"
	ScenarioBackground  Scenario = "background"
	ScenarioThink       Scenario = "think"
	ScenarioComplex     Scenario = "complex"
	ScenarioLongContext Scenario = "long_context"
	ScenarioFast        Scenario = "fast"
)

// ScenarioResult contains the detected scenario and token count.
type ScenarioResult struct {
	Scenario   Scenario
	TokenCount int
	Reason     string
}

// MessageContent represents a single message in a conversation.
type MessageContent struct {
	Role    string
	Content string
}

// DetectScenario analyzes a request to determine which model to use.
// Routing priority:
//  1. Long Context (> threshold) - 硬性限制，最先判断
//  2. Model Tier (sonnet -> think, haiku -> fast)
//  3. Complex (architectural patterns or tool-heavy operations)
//  4. Think (reasoning patterns)
//  5. Background (simple operations with NO tools)
//  6. Default
//
// For streaming requests, consider using RouteForStreaming() to prefer faster models.
func DetectScenario(messages []MessageContent, tokenCount int, cfg *config.Config, modelID string) ScenarioResult {
	// 1. Check for long context first (most important - hard constraint)
	threshold := getLongContextThreshold(cfg)
	if tokenCount > threshold {
		return ScenarioResult{
			Scenario:   ScenarioLongContext,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("token count %d exceeds threshold %d (Recommand MiniMax for 1M context)", tokenCount, threshold),
		}
	}

	// 2. Check model tier (sonnet -> think, haiku -> fast)
	if tier := getModelTier(modelID); tier != "" {
		switch tier {
		case "sonnet":
			return ScenarioResult{
				Scenario:   ScenarioThink,
				TokenCount: tokenCount,
				Reason:     fmt.Sprintf("model tier sonnet -> think scenario (model: %s)", modelID),
			}
		case "haiku":
			return ScenarioResult{
				Scenario:   ScenarioFast,
				TokenCount: tokenCount,
				Reason:     fmt.Sprintf("model tier haiku -> fast scenario (model: %s)", modelID),
			}
		}
	}

	// 3. Check for complex tasks (architectural OR tool-related)
	if hasComplexPattern(messages) {
		return ScenarioResult{
			Scenario:   ScenarioComplex,
			TokenCount: tokenCount,
			Reason:     "complex or tool-based operation detected (Recommand GLM-5.1)",
		}
	}

	// 4. Check for thinking/reasoning patterns
	if hasThinkingPattern(messages) {
		return ScenarioResult{
			Scenario:   ScenarioThink,
			TokenCount: tokenCount,
			Reason:     "thinking/reasoning pattern detected (Recommand GLM-5)",
		}
	}

	// 5. Check for background task patterns (truly simple operations)
	if hasBackgroundPattern(messages) {
		return ScenarioResult{
			Scenario:   ScenarioBackground,
			TokenCount: tokenCount,
			Reason:     "simple background task detected (Recommand Qwen3.5 Plus)",
		}
	}

	// 6. Default
	return ScenarioResult{
		Scenario:   ScenarioDefault,
		TokenCount: tokenCount,
		Reason:     "default scenario (Recommand Kimi K2.6)",
	}
}

// hasComplexPattern looks for complex operations that need more capable models.
// This includes tool-based operations (executing functions, writing/editing files, etc.)
func hasComplexPattern(messages []MessageContent) bool {
	complexKeywords := []string{
		// Architectural
		"architect", "architecture", "refactor", "redesign",
		"complex", "difficult", "challenging",
		"optimize", "performance", "efficiency",
		"design pattern", "best practice",
		// Tool-related keywords indicate complex operations
		"execute", "run command", "bash", "shell",
		"implement", "build", "create", "add feature",
		"write to", "edit file", "create file",
	}

	for _, msg := range messages {
		if msg.Role == "system" || msg.Role == "user" {
			lower := strings.ToLower(msg.Content)
			for _, kw := range complexKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
		}
	}
	return false
}

// hasThinkingPattern looks for system prompts mentioning reasoning keywords
// or content containing thinking/reasoning markers.
func hasThinkingPattern(messages []MessageContent) bool {
	thinkingKeywords := []string{
		"think", "thinking", "plan", "reason", "reasoning",
		"analyze", "analysis", "step by step",
	}

	for _, msg := range messages {
		if msg.Role == "system" || msg.Role == "user" {
			lower := strings.ToLower(msg.Content)
			for _, kw := range thinkingKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
		}
		// Check for thinking content blocks
		if strings.Contains(msg.Content, "antThinking") {
			return true
		}
	}
	return false
}

// hasBackgroundPattern checks for VERY simple background tasks.
// IMPORTANT: This should be conservative - returns true only for truly trivial requests.
// If there's any mention of tools, functions, or writing, it's NOT background.
func hasBackgroundPattern(messages []MessageContent) bool {
	// If ANY tool keywords appear, it's NOT a background task
	toolBlockers := []string{
		"tool", "function", "execute", "run command",
		"write", "edit", "create", "delete", "remove",
		"implement", "build", "add", "modify",
	}

	for _, msg := range messages {
		lower := strings.ToLower(msg.Content)
		for _, kw := range toolBlockers {
			if strings.Contains(lower, kw) {
				return false
			}
		}
	}

	// Only truly simple operations are background tasks
	backgroundKeywords := []string{
		"list directory", "ls -", "dir",
		"show file", "view file", "cat file",
		"what is", "what's", "tell me about",
		"check status", "show status",
	}

	for _, msg := range messages {
		lower := strings.ToLower(msg.Content)
		for _, kw := range backgroundKeywords {
			if strings.Contains(lower, kw) {
				return true
			}
		}
	}
	return false
}

// getLongContextThreshold returns the configured threshold or a sensible default.
// Default is 100K tokens to trigger long-context models (1M context) vs regular models (128-256K).
func getLongContextThreshold(cfg *config.Config) int {
	if cfg == nil {
		return 100000 // Default: 100K tokens
	}
	if lc, ok := cfg.Models["long_context"]; ok && lc.ContextThreshold > 0 {
		return lc.ContextThreshold
	}
	return 100000 // Default: 100K tokens
}

// getModelTier extracts the model tier from a model ID.
// Model ID format: claude-{tier}-{version}, e.g., claude-sonnet-5-7, claude-haiku-4-7
// Returns: "sonnet", "haiku", "opus", or "" if unknown
func getModelTier(modelID string) string {
	modelID = strings.ToLower(modelID)
	if strings.Contains(modelID, "sonnet") {
		return "sonnet"
	}
	if strings.Contains(modelID, "haiku") {
		return "haiku"
	}
	return ""
}

// RouteForStreaming selects a model optimized for streaming latency.
// For streaming, we prioritize fast TTFT (time-to-first-token) over capability.
// This may return a less capable model but one that streams faster.
func RouteForStreaming(messages []MessageContent, tokenCount int, cfg *config.Config) ScenarioResult {
	// For streaming, use simpler models that have better TTFT
	// Complex models (GLM, Kimi) are too slow for streaming with many tools

	threshold := getLongContextThreshold(cfg)
	if tokenCount > threshold {
		model := "long_context"
		if cfg != nil {
			if lc, ok := cfg.Models["long_context"]; ok && lc.ModelID != "" {
				model = lc.ModelID
			}
		}
		return ScenarioResult{
			Scenario:   ScenarioLongContext,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("high token count streaming (%d > %d) - use %s for acceptable TTFT", tokenCount, threshold, model),
		}
	}

	if hasComplexPattern(messages) || hasThinkingPattern(messages) {
		// Complex request but streaming - downgrade to faster model
		// GLM-5 and Kimi are too slow for streaming with complex prompts
		return ScenarioResult{
			Scenario:   ScenarioFast,
			TokenCount: tokenCount,
			Reason:     "complex request but streaming - use fast model (qwen3.6-plus) for better TTFT",
		}
	}

	// Default to fast scenario for streaming
	return ScenarioResult{
		Scenario:   ScenarioFast,
		TokenCount: tokenCount,
		Reason:     "streaming request - use fast model (qwen3.6-plus)",
	}
}
