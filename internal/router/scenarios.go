package router

import (
	"fmt"
	"regexp"
	"strings"

	"oc-go-cc/internal/complexity"
	"oc-go-cc/internal/config"
)

var systemReminderRe = regexp.MustCompile(`(?s)<system-reminder>.*?</system-reminder>`)

func stripSystemReminders(text string) string {
	return strings.TrimSpace(systemReminderRe.ReplaceAllString(text, ""))
}

type Scenario string

const (
	ScenarioDefault     Scenario = "default"
	ScenarioBackground Scenario = "background"
	ScenarioThink     Scenario = "think"
	ScenarioComplex    Scenario = "complex"
	ScenarioLongContext Scenario = "long_context"
	ScenarioFast      Scenario = "fast"
)

type ScenarioResult struct {
	Scenario   Scenario
	TokenCount int
	Reason    string
}

type MessageContent struct {
	Role    string
	Content string
}

func DetectScenario(messages []MessageContent, tokenCount int, cfg *config.Config, modelID string) ScenarioResult {
	threshold := getLongContextThreshold(cfg)
	if tokenCount > threshold {
		return ScenarioResult{
			Scenario:   ScenarioLongContext,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("token count %d exceeds threshold %d (Recommand MiniMax for 1M context)", tokenCount, threshold),
		}
	}

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

	return ScenarioResult{
		Scenario:   ScenarioDefault,
		TokenCount: tokenCount,
		Reason:     "default scenario (Recommand Kimi K2.6)",
	}
}

func DetectScenarioWithComplexity(compReq *complexity.Request, cfg *config.Config, modelID string) ScenarioResult {
	threshold := getLongContextThreshold(cfg)
	if compReq == nil {
		return ScenarioResult{
			Scenario:   ScenarioFast,
			TokenCount: 0,
			Reason:     "nil request, fast scenario",
		}
	}

	result, err := complexity.Analyze(compReq)
	if err != nil {
		return ScenarioResult{
			Scenario:   ScenarioFast,
			TokenCount: 0,
			Reason:     fmt.Sprintf("complexity analysis failed: %v, fast scenario", err),
		}
	}

	tokenCount := result.Metrics.TotalTokens

	if tokenCount > threshold {
		return ScenarioResult{
			Scenario:   ScenarioLongContext,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("token count %d exceeds threshold %d (Recommand MiniMax for 1M context)", tokenCount, threshold),
		}
	}

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

	if complexity.HasWriteToolsInCurrentUserMessage(compReq) {
		switch result.Level {
		case complexity.LevelExtreme:
			return ScenarioResult{
				Scenario:   ScenarioComplex,
				TokenCount: tokenCount,
				Reason:     fmt.Sprintf("complexity score %d (extreme): %s (Recommand GLM-5.1)", result.TotalScore, summarizeDimensions(result)),
			}
		case complexity.LevelComplex:
			return ScenarioResult{
				Scenario:   ScenarioThink,
				TokenCount: tokenCount,
				Reason:     fmt.Sprintf("complexity score %d (complex): %s (Recommand GLM-5)", result.TotalScore, summarizeDimensions(result)),
			}
		case complexity.LevelMedium:
			return ScenarioResult{
				Scenario:   ScenarioDefault,
				TokenCount: tokenCount,
				Reason:     fmt.Sprintf("complexity score %d (medium): %s (Recommand Kimi K2.6)", result.TotalScore, summarizeDimensions(result)),
			}
		default:
			return ScenarioResult{
				Scenario:   ScenarioDefault,
				TokenCount: tokenCount,
				Reason:     fmt.Sprintf("complexity score %d (simple): has write tools (Recommand Kimi K2.6)", result.TotalScore),
			}
		}
	}

	if hasComplexKeywordsInCurrentMessage(compReq) {
		switch result.Level {
		case complexity.LevelExtreme:
			return ScenarioResult{
				Scenario:   ScenarioComplex,
				TokenCount: tokenCount,
				Reason:     fmt.Sprintf("complexity score %d (extreme): %s (Recommand GLM-5.1)", result.TotalScore, summarizeDimensions(result)),
			}
		case complexity.LevelComplex:
			return ScenarioResult{
				Scenario:   ScenarioThink,
				TokenCount: tokenCount,
				Reason:     fmt.Sprintf("complexity score %d (complex): %s (Recommand GLM-5)", result.TotalScore, summarizeDimensions(result)),
			}
		default:
			return ScenarioResult{
				Scenario:   ScenarioThink,
				TokenCount: tokenCount,
				Reason:     fmt.Sprintf("complexity score %d (medium): has complex keywords (Recommand GLM-5)", result.TotalScore),
			}
		}
	}

	if hasActualSimpleToolCalls(compReq) || hasSimpleMessagePattern(compReq) {
		return ScenarioResult{
			Scenario:   ScenarioFast,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("simple read-only operation detected (Recommand Qwen3.5 Plus)"),
		}
	}

	switch result.Level {
	case complexity.LevelExtreme:
		return ScenarioResult{
			Scenario:   ScenarioComplex,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("complexity score %d (extreme): %s (Recommand GLM-5.1)", result.TotalScore, summarizeDimensions(result)),
		}
	case complexity.LevelComplex:
		return ScenarioResult{
			Scenario:   ScenarioThink,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("complexity score %d (complex): %s (Recommand GLM-5)", result.TotalScore, summarizeDimensions(result)),
		}
	case complexity.LevelMedium:
		return ScenarioResult{
			Scenario:   ScenarioDefault,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("complexity score %d (medium): %s (Recommand Kimi K2.6)", result.TotalScore, summarizeDimensions(result)),
		}
	default:
		return ScenarioResult{
			Scenario:   ScenarioFast,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("complexity score %d (simple): fast scenario (Recommand Qwen3.5 Plus)", result.TotalScore),
		}
	}
}

func summarizeDimensions(r *complexity.Result) string {
	parts := []string{}
	for _, d := range r.Dimensions {
		if d.Score > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d/%d", d.Name, d.Score, d.MaxScore))
		}
	}
	if len(parts) == 0 {
		return "no pressure"
	}
	return strings.Join(parts, ", ")
}

func getLongContextThreshold(cfg *config.Config) int {
	if cfg == nil {
		return 100000
	}
	if lc, ok := cfg.Models["long_context"]; ok && lc.ContextThreshold > 0 {
		return lc.ContextThreshold
	}
	return 100000
}

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

var simpleReadOnlyTools = []string{
	"explore",
	"search",
	"grep",
	"read",
	"view",
	"cat",
	"ls",
	"list",
	"find",
	"fetch",
	"query",
	"inspect",
	"glob",
}

func isSimpleToolName(toolName, pattern string) bool {
	lower := strings.ToLower(toolName)
	return lower == pattern || strings.HasPrefix(lower, pattern+"_") || strings.HasPrefix(lower, pattern+"-")
}

var simpleMessagePatterns = []string{
	"list directory",
	"ls -la",
	"ls -l",
	"ls -",
	"show file",
	"view file",
	"cat file",
	"cat /",
	"cat ./",
	"what is this",
	"what's this",
	"tell me about",
	"check status",
	"show status",
}

var complexKeywords = []string{
	"architect", "architecture", "refactor", "redesign",
	"complex", "challenging",
	"optimize", "performance", "efficiency",
	"design pattern", "best practice",
	"execute", "bash", "shell",
	"create",
	"write to", "edit file", "create file",
	"重构", "重写", "重新设计",
	"架构", "架构设计", "系统设计",
	"优化", "性能优化", "提升性能",
	"复杂", "挑战",
	"最佳实践", "设计模式",
	"执行", "运行脚本", "运行命令",
	"部署",
}

func isSimpleToolOnly(req *complexity.Request) bool {
	if req == nil || len(req.Tools) == 0 {
		return false
	}
	for _, tool := range req.Tools {
		for _, t := range simpleReadOnlyTools {
			if isSimpleToolName(tool.Name, t) {
				return true
			}
		}
	}
	return false
}

func hasActualSimpleToolCalls(req *complexity.Request) bool {
	if req == nil {
		return false
	}
	for i := len(req.Messages) - 1; i >= 0; i-- {
		msg := req.Messages[i]
		if msg.Role == "user" {
			for _, block := range msg.Blocks {
				if block.Type == "tool_use" {
					for _, t := range simpleReadOnlyTools {
						if isSimpleToolName(block.Name, t) {
							return true
						}
					}
				}
			}
			return false
		}
	}
	return false
}

func hasSimpleMessagePattern(req *complexity.Request) bool {
	if req == nil {
		return false
	}
	for i := len(req.Messages) - 1; i >= 0; i-- {
		msg := req.Messages[i]
		if msg.Role == "user" {
			text := strings.ToLower(stripSystemReminders(msg.Content))
			for _, pattern := range simpleMessagePatterns {
				if strings.Contains(text, pattern) {
					return true
				}
			}
			return false
		}
	}
	return false
}

func hasComplexKeywordsInCurrentMessage(req *complexity.Request) bool {
	if req == nil {
		return false
	}
	for i := len(req.Messages) - 1; i >= 0; i-- {
		msg := req.Messages[i]
		if msg.Role == "user" {
			text := strings.ToLower(stripSystemReminders(msg.Content))
			for _, keyword := range complexKeywords {
				if strings.Contains(text, keyword) {
					return true
				}
			}
			return false
		}
	}
	return false
}

func extractAllTexts(req *complexity.Request) []string {
	var texts []string
	for _, s := range req.System {
		if s.Text != "" {
			texts = append(texts, s.Text)
		}
	}
	for _, msg := range req.Messages {
		if msg.Content != "" {
			texts = append(texts, msg.Content)
		}
	}
	return texts
}

func RouteForStreaming(messages []MessageContent, tokenCount int, cfg *config.Config) ScenarioResult {
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

	return ScenarioResult{
		Scenario:   ScenarioFast,
		TokenCount: tokenCount,
		Reason:     "streaming request - use fast model (qwen3.6-plus)",
	}
}

func RouteForStreamingWithComplexity(compReq *complexity.Request, cfg *config.Config) ScenarioResult {
	if compReq == nil {
		return ScenarioResult{
			Scenario:   ScenarioFast,
			TokenCount: 0,
			Reason:     "nil request, fast scenario for streaming",
		}
	}

	result, err := complexity.Analyze(compReq)
	if err != nil {
		return ScenarioResult{
			Scenario:   ScenarioFast,
			TokenCount: 0,
			Reason:     fmt.Sprintf("complexity analysis failed: %v, fast scenario for streaming", err),
		}
	}

	tokenCount := result.Metrics.TotalTokens
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

	return ScenarioResult{
		Scenario:   ScenarioFast,
		TokenCount: tokenCount,
		Reason:     fmt.Sprintf("streaming request (complexity: %d/%s) - use fast model (qwen3.6-plus)", result.TotalScore, result.Level),
	}
}
