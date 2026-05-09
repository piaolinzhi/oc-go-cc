package complexity

import (
	"encoding/json"
	"fmt"
	"strings"
)

const DefaultContextWindow = 200000

type SystemBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Message struct {
	Role    string       `json:"role"`
	Content string      `json:"content"`
	Blocks  []ContentBlock `json:"blocks,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
	Text string `json:"text,omitempty"`
}

type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Request struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	System    []SystemBlock   `json:"system,omitempty"`
	Messages  []Message      `json:"messages"`
	Tools     []ToolDefinition `json:"tools,omitempty"`
	Stream    *bool          `json:"stream,omitempty"`
}

func ParseRequest(data []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}
	return &req, nil
}

func extractSystemTexts(req *Request) []string {
	texts := make([]string, 0, len(req.System))
	for _, s := range req.System {
		if s.Text != "" {
			texts = append(texts, s.Text)
		}
	}
	return texts
}

func extractAllTexts(req *Request) []string {
	texts := extractSystemTexts(req)
	for _, msg := range req.Messages {
		if msg.Content != "" {
			texts = append(texts, msg.Content)
		}
	}
	return texts
}

func countMemories(req *Request) int {
	count := 0
	for _, msg := range req.Messages {
		lower := strings.ToLower(msg.Content)
		if strings.Contains(lower, "system-reminder") ||
			strings.Contains(lower, "session_start") ||
			strings.Contains(lower, "sessionstart") ||
			strings.Contains(lower, "memory") ||
			strings.Contains(lower, "记忆") {
			count++
		}
	}
	return count
}

func countActualToolCalls(req *Request) int {
	count := 0
	for _, msg := range req.Messages {
		for _, block := range msg.Blocks {
			if block.Type == "tool_use" {
				count++
			}
		}
	}
	return count
}

func HasActualWriteTools(req *Request) bool {
	for _, msg := range req.Messages {
		for _, block := range msg.Blocks {
			if block.Type == "tool_use" {
				lower := strings.ToLower(block.Name)
				for _, tool := range []string{"write", "edit", "delete", "remove", "execute", "run", "bash", "shell", "create", "modify", "update"} {
					if strings.Contains(lower, tool) {
						return true
					}
				}
			}
		}
	}
	return false
}

func HasWriteToolsInCurrentUserMessage(req *Request) bool {
	for i := len(req.Messages) - 1; i >= 0; i-- {
		msg := req.Messages[i]
		if msg.Role == "user" {
			for _, block := range msg.Blocks {
				if block.Type == "tool_use" {
					lower := strings.ToLower(block.Name)
					for _, tool := range []string{"write", "edit", "delete", "remove", "execute", "run", "bash", "shell", "create", "modify", "update"} {
						if strings.Contains(lower, tool) {
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

func countFileReferences(req *Request) (fileCount int, fileLines int) {
	filePatterns := []string{
		"file_path", "filepath",
		"readfile", "writefile", "editfile",
		"cat ", "vim ", "nano ",
		".go\"", ".py\"", ".ts\"", ".js\"", ".java\"",
		".json\"", ".yaml\"", ".yml\"", ".md\"",
	}
	linePatterns := []string{
		"lines ", "line ", "row ",
		"行", "lines:",
	}

	allText := strings.ToLower(strings.Join(extractAllTexts(req), " "))

	for _, p := range filePatterns {
		count := strings.Count(allText, p)
		if count > 0 {
			fileCount += count
		}
	}

	for _, p := range linePatterns {
		count := strings.Count(allText, p)
		if count > 0 {
			fileLines += count * 10
		}
	}

	return fileCount, fileLines
}

func Analyze(req *Request) (*Result, error) {
	tc, err := NewTokenCounter()
	if err != nil {
		return nil, fmt.Errorf("failed to create token counter: %w", err)
	}

	systemTexts := extractSystemTexts(req)
	allTexts := extractAllTexts(req)

	totalTokens := tc.CountMessages(systemTexts, req.Messages)
	contextWindow := DefaultContextWindow

	constraintCount := countConstraints(allTexts)
	memoryCount := countMemories(req)
	fileCount, fileLines := countFileReferences(req)

	conversationTurns := 0
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			conversationTurns++
		}
	}
	toolCount := countActualToolCalls(req)

	logicSteps := countLogicSteps(allTexts)

	dimensions := []DimensionScore{
		scoreTokenUsage(totalTokens, contextWindow),
		scoreConstraints(constraintCount),
		scoreMemory(memoryCount),
		scoreFiles(fileCount, fileLines),
		scoreConversation(conversationTurns, toolCount),
		scoreLogicSteps(logicSteps),
	}

	totalScore := 0
	for _, d := range dimensions {
		totalScore += d.Score
	}

	pct := 0
	if contextWindow > 0 {
		pct = totalTokens * 100 / contextWindow
	}

	return &Result{
		Level:      levelFromScore(totalScore),
		TotalScore: totalScore,
		MaxScore:   100,
		Dimensions: dimensions,
		Metrics: Metrics{
			TotalTokens:        totalTokens,
			ContextWindow:      contextWindow,
			TokenUsagePct:      pct,
			ConstraintCount:    constraintCount,
			MemoryCount:        memoryCount,
			FileCount:          fileCount,
			FileLines:          fileLines,
			ConversationTurns: conversationTurns,
			ToolCount:          toolCount,
			LogicSteps:         logicSteps,
		},
	}, nil
}

func AnalyzeRaw(data []byte) (*Result, error) {
	req, err := ParseRequest(data)
	if err != nil {
		return nil, err
	}
	return Analyze(req)
}
