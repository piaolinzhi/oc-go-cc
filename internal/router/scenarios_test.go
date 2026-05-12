package router

import (
	"strings"
	"testing"

	"oc-go-cc/internal/complexity"
	"oc-go-cc/internal/config"
)

func mockConfig() *config.Config {
	return &config.Config{
		Models: map[string]config.ModelConfig{
			"long_context": {
				ContextThreshold: 60000,
			},
		},
	}
}

func TestDetectScenarioWithComplexity_Simple(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "Hello, how are you?"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for simple request, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestDetectScenarioWithComplexity_Complex(t *testing.T) {
	cfg := &config.Config{
		Models: map[string]config.ModelConfig{
			"long_context": {
				ContextThreshold: 1000000,
			},
		},
	}
	longText := ""
	for i := 0; i < 15000; i++ {
		longText += "你必须按照规范实现，禁止使用违规方法，必须遵循最佳实践，必须保证代码质量。"
	}
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		System: []complexity.SystemBlock{
			{Type: "text", Text: longText},
		},
		Messages: []complexity.Message{
			{Role: "user", Content: "system-reminder: memory item 1"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 2"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 3"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 4"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 5"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 6"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 7"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 8"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "step 1. step 2. step 3. step 4. step 5. step 6. step 7. step 8. step 9. step 10. step 11. step 12. step 13.", Blocks: []complexity.ContentBlock{
				{Type: "tool_use", Name: "write_file", Text: "create new file"},
			}},
		},
		Tools: []complexity.ToolDefinition{
			{Name: "read_file", Description: "read"},
			{Name: "write_file", Description: "write"},
			{Name: "edit_file", Description: "edit"},
			{Name: "search", Description: "search"},
			{Name: "execute", Description: "exec"},
			{Name: "deploy", Description: "deploy"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, cfg, "")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for complex request with write tool call, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestDetectScenarioWithComplexity_LongContextPriority(t *testing.T) {
	longText := ""
	for i := 0; i < 5000; i++ {
		longText += "这是一段很长的文本用来测试长上下文检测。"
	}
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		System: []complexity.SystemBlock{
			{Type: "text", Text: longText},
		},
		Messages: []complexity.Message{
			{Role: "user", Content: "hello"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioLongContext {
		t.Errorf("Expected ScenarioLongContext, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestDetectScenarioWithComplexity_ModelTierSonnet(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "hello"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "claude-sonnet-5-7")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for sonnet model, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestDetectScenarioWithComplexity_ModelTierHaiku(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-haiku-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "hello"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "claude-haiku-4-7")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for haiku model, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestDetectScenarioWithComplexity_NilRequest(t *testing.T) {
	result := DetectScenarioWithComplexity(nil, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for nil request, got %s", result.Scenario)
	}
}

func TestDetectScenarioWithComplexity_Medium(t *testing.T) {
	longText := ""
	for i := 0; i < 2000; i++ {
		longText += "你必须按照规范实现，保证兼容性，禁止违反架构规范。"
	}
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		System: []complexity.SystemBlock{
			{Type: "text", Text: longText},
		},
		Messages: []complexity.Message{
			{Role: "user", Content: "帮我写一个函数"},
			{Role: "assistant", Content: "好的"},
			{Role: "user", Content: "还需要处理错误"},
			{Role: "assistant", Content: "已添加"},
			{Role: "user", Content: "加上单元测试"},
			{Role: "user", Content: "再加一个测试用例"},
		},
		Tools: []complexity.ToolDefinition{
			{Name: "read_file", Description: "read"},
			{Name: "write_file", Description: "write"},
			{Name: "edit_file", Description: "edit"},
			{Name: "search", Description: "search"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario == ScenarioDefault || result.Scenario == ScenarioBackground {
		t.Errorf("Expected non-trivial scenario for medium request, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestDetectScenario_OldAPI(t *testing.T) {
	messages := []MessageContent{
		{Role: "user", Content: "Hello"},
	}
	result := DetectScenario(messages, 100, mockConfig(), "")
	if result.Scenario != ScenarioDefault {
		t.Errorf("Expected ScenarioDefault, got %s", result.Scenario)
	}
}

func TestDetectScenario_OldAPI_LongContext(t *testing.T) {
	messages := []MessageContent{
		{Role: "user", Content: "Refactor this code"},
	}
	result := DetectScenario(messages, 70000, mockConfig(), "")
	if result.Scenario != ScenarioLongContext {
		t.Errorf("Expected ScenarioLongContext, got %s", result.Scenario)
	}
}

func TestRouteForStreaming_RespectsConfiguredThreshold(t *testing.T) {
	messages := []MessageContent{
		{Role: "user", Content: "Hello"},
	}
	cfg := &config.Config{
		Models: map[string]config.ModelConfig{
			"long_context": {
				ModelID:          "deepseek-v4-flash",
				ContextThreshold: 256000,
			},
		},
	}

	result := RouteForStreaming(messages, 40955, cfg)
	if result.Scenario == ScenarioLongContext {
		t.Errorf("Expected NOT ScenarioLongContext for 40955 tokens with threshold 256000, got %s", result.Scenario)
	}

	result = RouteForStreaming(messages, 300000, cfg)
	if result.Scenario != ScenarioLongContext {
		t.Errorf("Expected ScenarioLongContext for 300000 tokens with threshold 256000, got %s", result.Scenario)
	}
	if !strings.Contains(result.Reason, "deepseek-v4-flash") {
		t.Errorf("Expected reason to mention configured model 'deepseek-v4-flash', got: %s", result.Reason)
	}
}

func TestRouteForStreaming_NilConfig(t *testing.T) {
	messages := []MessageContent{
		{Role: "user", Content: "Hello"},
	}

	result := RouteForStreaming(messages, 90000, nil)
	if result.Scenario == ScenarioLongContext {
		t.Errorf("Expected NOT ScenarioLongContext for 90000 tokens with nil config, got %s", result.Scenario)
	}

	result = RouteForStreaming(messages, 110000, nil)
	if result.Scenario != ScenarioLongContext {
		t.Errorf("Expected ScenarioLongContext for 110000 tokens with nil config, got %s", result.Scenario)
	}
}

func TestRouteForStreamingWithComplexity(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "hello"},
		},
	}
	result := RouteForStreamingWithComplexity(compReq, mockConfig())
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for streaming simple request, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestRouteForStreamingWithComplexity_LongContext(t *testing.T) {
	longText := ""
	for i := 0; i < 5000; i++ {
		longText += "这是一段很长的文本用来测试长上下文检测。"
	}
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		System: []complexity.SystemBlock{
			{Type: "text", Text: longText},
		},
		Messages: []complexity.Message{
			{Role: "user", Content: "hello"},
		},
	}
	result := RouteForStreamingWithComplexity(compReq, mockConfig())
	if result.Scenario != ScenarioLongContext {
		t.Errorf("Expected ScenarioLongContext for streaming long request, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestRouteForStreamingWithComplexity_NilRequest(t *testing.T) {
	result := RouteForStreamingWithComplexity(nil, mockConfig())
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for nil request, got %s", result.Scenario)
	}
}

func TestIsSimpleToolOnly_Explore(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "explore the codebase"},
		},
		Tools: []complexity.ToolDefinition{
			{Name: "explore", Description: "explore files"},
			{Name: "search", Description: "search content"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for explore/search tools, got %s", result.Scenario)
	}
}

func TestIsSimpleToolOnly_MixedTools(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "read and write files", Blocks: []complexity.ContentBlock{
				{Type: "tool_use", Name: "write_file", Text: "write to file"},
			}},
		},
		Tools: []complexity.ToolDefinition{
			{Name: "read_file", Description: "read"},
			{Name: "write_file", Description: "write"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario == ScenarioFast {
		t.Errorf("Expected NOT ScenarioFast for current message with write tool call, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestIsSimpleToolOnly_NoTools(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "hello"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for hello, got %s", result.Scenario)
	}
}

func TestHasSimpleMessagePattern_ListDirectory(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "list directory contents"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for list directory, got %s", result.Scenario)
	}
}

func TestHasSimpleMessagePattern_WhatIs(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "what is this function doing?"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for what is question, got %s", result.Scenario)
	}
}

func TestHasSimpleMessagePattern_CatFile(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "cat file /path/to/config"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for cat file, got %s", result.Scenario)
	}
}

func TestHasComplexKeywords_Refactor(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "refactor this function"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for refactor, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestHasComplexKeywords_Execute(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "execute this script"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for execute, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestHasComplexKeywords_Architecture(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "improve the architecture design"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for architecture, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestHasComplexKeywords_Performance(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "optimize the performance"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for performance, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestHasComplexKeywords_Chinese_Refactor(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "帮我重构这个函数"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for Chinese refactor, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestHasComplexKeywords_Chinese_Architecture(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "优化系统架构"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for Chinese architecture, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestHasComplexKeywords_Chinese_Performance(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "提升性能"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for Chinese performance, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestHasComplexKeywords_Chinese_Deploy(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "部署应用到服务器"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioThink {
		t.Errorf("Expected ScenarioThink for Chinese deploy, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestStripSystemReminders_Basic(t *testing.T) {
	input := "<system-reminder>create a file and execute bash shell</system-reminder>\nhello"
	got := stripSystemReminders(input)
	if strings.Contains(got, "system-reminder") {
		t.Errorf("Expected system-reminder to be stripped, got: %s", got)
	}
	if !strings.Contains(got, "hello") {
		t.Errorf("Expected 'hello' to remain, got: %s", got)
	}
}

func TestStripSystemReminders_Multiple(t *testing.T) {
	input := "<system-reminder>first\ncreate\necho\n</system-reminder>hello<system-reminder>second\nexecute\nshell\n</system-reminder>"
	got := stripSystemReminders(input)
	if strings.Contains(got, "system-reminder") {
		t.Errorf("Expected all system-reminders to be stripped, got: %s", got)
	}
	if !strings.Contains(got, "hello") {
		t.Errorf("Expected 'hello' to remain, got: %s", got)
	}
}

func TestStripSystemReminders_NoReminders(t *testing.T) {
	input := "hello world"
	got := stripSystemReminders(input)
	if got != "hello world" {
		t.Errorf("Expected unchanged text, got: %s", got)
	}
}

func TestHelloWithSystemReminder_NotComplex(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "<system-reminder>\nYou have superpowers.\n\n**Below is the full content of your 'using-superpowers' skill**\n\nWhen you think there is even a 1% chance a skill might apply, you ABSOLUTELY MUST invoke the Skill tool.\n\nTools: Bash, Shell, Execute, Create file\n</system-reminder>\n\nhello"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for hello with system-reminder, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestHelloWithSystemReminder_NotSimplePattern(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "<system-reminder>\nUse cat /etc/hosts to check. ls -la to list files.\n</system-reminder>\n\nhello"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for hello with system-reminder containing simple patterns, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestRealHello_IsFast(t *testing.T) {
	compReq := &complexity.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []complexity.Message{
			{Role: "user", Content: "hello"},
		},
	}
	result := DetectScenarioWithComplexity(compReq, mockConfig(), "")
	if result.Scenario != ScenarioFast {
		t.Errorf("Expected ScenarioFast for plain hello, got %s (reason: %s)", result.Scenario, result.Reason)
	}
}

func TestIsSimpleToolName(t *testing.T) {
	tests := []struct {
		toolName string
		pattern  string
		want     bool
	}{
		{"read", "read", true},
		{"Read", "read", true},
		{"read_file", "read", true},
		{"ReadFile", "read", false},
		{"search", "search", true},
		{"SearchCodebase", "search", false},
		{"get", "get", true},
		{"GetDiagnostics", "get", false},
		{"grep", "grep", true},
		{"Grep", "grep", true},
		{"explore", "explore", true},
		{"Explore", "explore", true},
	}
	for _, tt := range tests {
		got := isSimpleToolName(tt.toolName, tt.pattern)
		if got != tt.want {
			t.Errorf("isSimpleToolName(%q, %q) = %v, want %v", tt.toolName, tt.pattern, got, tt.want)
		}
	}
}
