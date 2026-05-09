package complexity

import (
	"regexp"
	"strings"
)

var constraintPatterns = []*regexp.Regexp{
	regexp.MustCompile(`必须`),
	regexp.MustCompile(`禁止`),
	regexp.MustCompile(`遵循`),
	regexp.MustCompile(`不得`),
	regexp.MustCompile(`保证`),
	regexp.MustCompile(`(?i)must\s+not\b`),
	regexp.MustCompile(`(?i)shall\s+not\b`),
	regexp.MustCompile(`(?i)forbidden\b`),
	regexp.MustCompile(`(?i)never\s+(use|modify|delete|execute|generate)`),
	regexp.MustCompile(`(?i)must\s+(implement|follow|use|adhere)`),
	regexp.MustCompile(`(?i)must\s+ensure\b`),
	regexp.MustCompile(`(?i)must\s+comply\b`),
}

func countConstraints(texts []string) int {
	count := 0
	for _, text := range texts {
		for _, p := range constraintPatterns {
			matches := p.FindAllString(text, -1)
			count += len(matches)
		}
	}
	return count
}

func scoreTokenUsage(totalTokens, contextWindow int) DimensionScore {
	pct := 0
	if contextWindow > 0 {
		pct = totalTokens * 100 / contextWindow
	}

	score := 0
	desc := ""
	switch {
	case pct > 70:
		score = 30
		desc = "token usage >70%, context nearly exhausted"
	case pct > 50:
		score = 20
		desc = "token usage 50%~70%, high context pressure"
	case pct > 25:
		score = 10
		desc = "token usage 25%~50%, moderate context load"
	default:
		desc = "token usage <25%, minimal context"
	}

	return DimensionScore{
		Name:        "token_usage",
		Weight:      30,
		Score:       score,
		MaxScore:    30,
		RawValue:    pct,
		Description: desc,
	}
}

func scoreConstraints(count int) DimensionScore {
	score := 0
	desc := ""
	switch {
	case count >= 15:
		score = 20
		desc = ">=15 constraints, heavily constrained"
	case count >= 8:
		score = 13
		desc = "8~14 constraints, moderately constrained"
	case count >= 4:
		score = 7
		desc = "4~7 constraints, lightly constrained"
	case count >= 1:
		score = 3
		desc = "1~3 constraints, minimal constraints"
	default:
		desc = "0 constraints, unconstrained"
	}

	return DimensionScore{
		Name:        "constraints",
		Weight:      20,
		Score:       score,
		MaxScore:    20,
		RawValue:    count,
		Description: desc,
	}
}

func scoreMemory(memoryCount int) DimensionScore {
	score := 0
	desc := ""
	switch {
	case memoryCount >= 8:
		score = 10
		desc = ">=8 memories, heavy memory load"
	case memoryCount >= 4:
		score = 6
		desc = "4~7 memories, moderate memory load"
	case memoryCount >= 1:
		score = 3
		desc = "1~3 memories, light memory load"
	default:
		desc = "no memory loaded"
	}

	return DimensionScore{
		Name:        "memory",
		Weight:      10,
		Score:       score,
		MaxScore:    10,
		RawValue:    memoryCount,
		Description: desc,
	}
}

func scoreFiles(fileCount, fileLines int) DimensionScore {
	score := 0
	desc := ""
	switch {
	case fileCount > 15 || fileLines > 800:
		score = 15
		desc = ">15 files or >800 lines, massive scope"
	case fileCount >= 8 || fileLines >= 400:
		score = 10
		desc = "8~15 files or 400~800 lines, large scope"
	case fileCount >= 3 || fileLines >= 100:
		score = 5
		desc = "3~7 files or 100~400 lines, moderate scope"
	default:
		desc = "<3 files and <100 lines, minimal scope"
	}

	return DimensionScore{
		Name:        "files",
		Weight:      15,
		Score:       score,
		MaxScore:    15,
		RawValue:    map[string]int{"file_count": fileCount, "file_lines": fileLines},
		Description: desc,
	}
}

func scoreConversation(turns, toolCount int) DimensionScore {
	score := 0
	desc := ""
	switch {
	case turns > 15:
		score = 10
		desc = ">15 turns, heavy interaction"
	case turns >= 10:
		score = 7
		desc = "10~15 turns, moderate interaction"
	case turns >= 5:
		score = 3
		desc = "5~9 turns, light interaction"
	default:
		desc = "<5 turns, minimal interaction"
	}

	return DimensionScore{
		Name:        "conversation",
		Weight:      10,
		Score:       score,
		MaxScore:    10,
		RawValue:    map[string]int{"turns": turns, "tool_count": toolCount},
		Description: desc,
	}
}

func scoreLogicSteps(steps int) DimensionScore {
	score := 0
	desc := ""
	switch {
	case steps > 12:
		score = 15
		desc = ">12 steps, extremely deep logic"
	case steps >= 8:
		score = 10
		desc = "8~12 steps, deep logic"
	case steps >= 4:
		score = 5
		desc = "4~7 steps, moderate logic"
	default:
		desc = "<4 steps, simple logic"
	}

	return DimensionScore{
		Name:        "logic_steps",
		Weight:      15,
		Score:       score,
		MaxScore:    15,
		RawValue:    steps,
		Description: desc,
	}
}

func countLogicSteps(texts []string) int {
	steps := 0
	stepIndicators := []string{
		"step 1", "step 2", "step 3", "step 4", "step 5",
		"步骤1", "步骤2", "步骤3", "步骤4", "步骤5",
		"first,", "second,", "third,", "then,", "finally,",
		"首先", "然后", "接着", "最后", "其次",
		"1.", "2.", "3.", "4.", "5.", "6.", "7.", "8.", "9.", "10.",
	}
	for _, text := range texts {
		lower := strings.ToLower(text)
		for _, indicator := range stepIndicators {
			count := strings.Count(lower, indicator)
			steps += count
		}
	}
	if steps == 0 {
		steps = 1
	}
	return steps
}
