package complexity

import (
	"encoding/json"
	"testing"
)

func TestAnalyzeSimple(t *testing.T) {
	req := &Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	}

	result, err := Analyze(req)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.Level != LevelSimple {
		t.Errorf("expected level simple, got %s (score: %d)", result.Level, result.TotalScore)
	}

	if result.TotalScore > 30 {
		t.Errorf("simple request should score <=30, got %d", result.TotalScore)
	}

	if result.Metrics.ConversationTurns != 1 {
		t.Errorf("expected 1 turn, got %d", result.Metrics.ConversationTurns)
	}

	if result.Metrics.ToolCount != 0 {
		t.Errorf("expected 0 tools, got %d", result.Metrics.ToolCount)
	}

	if result.Metrics.ConstraintCount != 0 {
		t.Errorf("expected 0 constraints, got %d", result.Metrics.ConstraintCount)
	}
}

func TestAnalyzeComplex(t *testing.T) {
	longText := ""
	for i := 0; i < 15000; i++ {
		longText += "你必须按照规范实现，禁止使用违规方法，必须遵循最佳实践，必须保证代码质量。"
	}

	req := &Request{
		Model: "claude-3-5-sonnet-20241022",
		System: []SystemBlock{
			{Type: "text", Text: longText},
		},
		Messages: []Message{
			{Role: "user", Content: "system-reminder: memory item 1"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 2"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 3"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "system-reminder: memory item 4"},
			{Role: "assistant", Content: "ok"},
			{Role: "user", Content: "step 1. step 2. step 3. step 4. step 5. step 6. step 7. step 8.", Blocks: []ContentBlock{
				{Type: "tool_use", Name: "read_file"},
				{Type: "tool_use", Name: "write_file"},
				{Type: "tool_use", Name: "edit_file"},
				{Type: "tool_use", Name: "search"},
				{Type: "tool_use", Name: "execute"},
				{Type: "tool_use", Name: "deploy"},
			}},
		},
		Tools: []ToolDefinition{
			{Name: "read_file", Description: "read"},
			{Name: "write_file", Description: "write"},
			{Name: "edit_file", Description: "edit"},
			{Name: "search", Description: "search"},
			{Name: "execute", Description: "exec"},
			{Name: "deploy", Description: "deploy"},
		},
	}

	result, err := Analyze(req)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.Level != LevelComplex {
		t.Errorf("expected level complex, got %s (score: %d)", result.Level, result.TotalScore)
	}

	if result.TotalScore <= 49 {
		t.Errorf("complex request should score >=50, got %d", result.TotalScore)
	}

	if result.Metrics.ConstraintCount == 0 {
		t.Error("expected constraints to be detected")
	}

	if result.Metrics.MemoryCount == 0 {
		t.Error("expected memories to be detected")
	}

	if result.Metrics.ToolCount != 6 {
		t.Errorf("expected 6 tools, got %d", result.Metrics.ToolCount)
	}
}

func TestAnalyzeRaw(t *testing.T) {
	payload := map[string]any{
		"model": "claude-3-5-sonnet-20241022",
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	result, err := AnalyzeRaw(data)
	if err != nil {
		t.Fatalf("AnalyzeRaw failed: %v", err)
	}

	if result.Level != LevelSimple {
		t.Errorf("expected simple, got %s", result.Level)
	}
}

func TestScoreTokenUsage(t *testing.T) {
	tests := []struct {
		tokens  int
		window  int
		expMin  int
	}{
		{30000, 200000, 0},
		{60000, 200000, 10},
		{110000, 200000, 20},
		{150000, 200000, 30},
	}

	for _, tt := range tests {
		score := scoreTokenUsage(tt.tokens, tt.window)
		if score.Score < tt.expMin {
			t.Errorf("tokens=%d window=%d: score %d < expected min %d", tt.tokens, tt.window, score.Score, tt.expMin)
		}
	}
}

func TestScoreConstraints(t *testing.T) {
	tests := []struct {
		count   int
		expMin  int
	}{
		{0, 0},
		{1, 3},
		{4, 7},
		{8, 13},
		{15, 20},
	}

	for _, tt := range tests {
		score := scoreConstraints(tt.count)
		if score.Score < tt.expMin {
			t.Errorf("count=%d: score %d < expected min %d", tt.count, score.Score, tt.expMin)
		}
	}
}

func TestScoreMemory(t *testing.T) {
	tests := []struct {
		count  int
		expMin int
	}{
		{0, 0},
		{1, 3},
		{4, 6},
		{8, 10},
	}

	for _, tt := range tests {
		score := scoreMemory(tt.count)
		if score.Score < tt.expMin {
			t.Errorf("count=%d: score %d < expected min %d", tt.count, score.Score, tt.expMin)
		}
	}
}

func TestScoreFiles(t *testing.T) {
	tests := []struct {
		fileCount int
		fileLines int
		expMin    int
	}{
		{1, 30, 0},
		{3, 100, 5},
		{5, 200, 5},
		{8, 400, 10},
		{16, 800, 15},
	}

	for _, tt := range tests {
		score := scoreFiles(tt.fileCount, tt.fileLines)
		if score.Score < tt.expMin {
			t.Errorf("files=%d lines=%d: score %d < expected min %d", tt.fileCount, tt.fileLines, score.Score, tt.expMin)
		}
	}
}

func TestScoreConversation(t *testing.T) {
	tests := []struct {
		turns     int
		toolCount int
		expMin    int
	}{
		{2, 0, 0},
		{3, 100, 0},
		{5, 100, 3},
		{8, 100, 3},
		{10, 100, 7},
		{12, 100, 7},
		{16, 100, 10},
	}

	for _, tt := range tests {
		score := scoreConversation(tt.turns, tt.toolCount)
		if score.Score < tt.expMin {
			t.Errorf("turns=%d tools=%d: score %d < expected min %d", tt.turns, tt.toolCount, score.Score, tt.expMin)
		}
	}
}

func TestScoreLogicSteps(t *testing.T) {
	tests := []struct {
		steps  int
		expMin int
	}{
		{2, 0},
		{4, 5},
		{8, 10},
		{13, 15},
	}

	for _, tt := range tests {
		score := scoreLogicSteps(tt.steps)
		if score.Score < tt.expMin {
			t.Errorf("steps=%d: score %d < expected min %d", tt.steps, score.Score, tt.expMin)
		}
	}
}

func TestLevelFromScore(t *testing.T) {
	tests := []struct {
		score int
		level Level
	}{
		{0, LevelSimple},
		{30, LevelSimple},
		{40, LevelSimple},
		{41, LevelMedium},
		{50, LevelMedium},
		{60, LevelMedium},
		{61, LevelComplex},
		{70, LevelComplex},
		{80, LevelComplex},
		{81, LevelExtreme},
		{100, LevelExtreme},
	}

	for _, tt := range tests {
		level := levelFromScore(tt.score)
		if level != tt.level {
			t.Errorf("score=%d: expected %s, got %s", tt.score, tt.level, level)
		}
	}
}

func TestCountConstraints(t *testing.T) {
	texts := []string{
		"必须 这样做",
		"禁止 做那件事",
		"必须 遵循规范",
		"禁止 违反规定",
		"必须 保证质量",
		"never 修改这个",
		"forbidden 这样做",
	}
	count := countConstraints(texts)
	if count < 5 {
		t.Errorf("expected >=5 constraints, got %d", count)
	}
}

func TestResultJSON(t *testing.T) {
	req := &Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	}

	result, err := Analyze(req)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var parsed Result
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if parsed.Level != result.Level {
		t.Errorf("level mismatch after JSON roundtrip: %s vs %s", parsed.Level, result.Level)
	}

	if parsed.TotalScore != result.TotalScore {
		t.Errorf("score mismatch after JSON roundtrip: %d vs %d", parsed.TotalScore, result.TotalScore)
	}
}
