package complexity

type Level string

const (
	LevelSimple  Level = "simple"
	LevelMedium Level = "medium"
	LevelComplex Level = "complex"
	LevelExtreme Level = "extreme"
)

type DimensionScore struct {
	Name        string `json:"name"`
	Weight      int    `json:"weight"`
	Score       int    `json:"score"`
	MaxScore    int    `json:"max_score"`
	RawValue    any    `json:"raw_value"`
	Description string `json:"description"`
}

type Metrics struct {
	TotalTokens      int `json:"total_tokens"`
	ContextWindow    int `json:"context_window"`
	TokenUsagePct    int `json:"token_usage_pct"`
	ConstraintCount  int `json:"constraint_count"`
	MemoryCount      int `json:"memory_count"`
	FileCount        int `json:"file_count"`
	FileLines        int `json:"file_lines"`
	ConversationTurns int `json:"conversation_turns"`
	ToolCount        int `json:"tool_count"`
	LogicSteps       int `json:"logic_steps"`
}

type Result struct {
	Level       Level            `json:"level"`
	TotalScore  int              `json:"total_score"`
	MaxScore    int              `json:"max_score"`
	Dimensions  []DimensionScore `json:"dimensions"`
	Metrics     Metrics          `json:"metrics"`
}

func levelFromScore(score int) Level {
	switch {
	case score <= 40:
		return LevelSimple
	case score <= 60:
		return LevelMedium
	case score <= 80:
		return LevelComplex
	default:
		return LevelExtreme
	}
}
