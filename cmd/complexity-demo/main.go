package main

import (
	"encoding/json"
	"fmt"
	"os"

	"oc-go-cc/internal/complexity"
)

func main() {
	examples := []struct {
		name string
		req  *complexity.Request
	}{
		{
			name: "Simple - hello",
			req: &complexity.Request{
				Model: "claude-3-5-sonnet-20241022",
				Messages: []complexity.Message{
					{Role: "user", Content: "hello"},
				},
			},
		},
		{
			name: "Medium - with constraints",
			req: &complexity.Request{
				Model: "claude-3-5-sonnet-20241022",
				System: []complexity.SystemBlock{
					{Type: "text", Text: "你必须按照规范实现，保证兼容性"},
				},
				Messages: []complexity.Message{
					{Role: "user", Content: "帮我写一个函数"},
					{Role: "assistant", Content: "好的"},
					{Role: "user", Content: "还需要处理错误"},
					{Role: "assistant", Content: "已添加"},
					{Role: "user", Content: "加上单元测试"},
				},
				Tools: []complexity.ToolDefinition{
					{Name: "read_file", Description: "read"},
				},
			},
		},
		{
			name: "Complex - multi-file with memory",
			req: &complexity.Request{
				Model: "claude-3-5-sonnet-20241022",
				System: []complexity.SystemBlock{
					{Type: "text", Text: "你必须按照架构规范实现，禁止违反规范，保证兼容性，遵循不得原则"},
				},
				Messages: []complexity.Message{
					{Role: "user", Content: "system-reminder: memory item 1"},
					{Role: "assistant", Content: "ok"},
					{Role: "user", Content: "system-reminder: memory item 2"},
					{Role: "assistant", Content: "ok"},
					{Role: "user", Content: "system-reminder: memory item 3"},
					{Role: "assistant", Content: "ok"},
					{Role: "user", Content: "请修改 file_path: main.go 和 file_path: handler.go"},
				},
				Tools: []complexity.ToolDefinition{
					{Name: "read_file", Description: "read"},
					{Name: "write_file", Description: "write"},
					{Name: "edit_file", Description: "edit"},
				},
			},
		},
		{
			name: "Extreme - full pressure",
			req: &complexity.Request{
				Model: "claude-3-5-sonnet-20241022",
				System: []complexity.SystemBlock{
					{Type: "text", Text: generateLongSystemPrompt()},
				},
				Messages: generateExtremeMessages(),
				Tools: []complexity.ToolDefinition{
					{Name: "read_file", Description: "read"},
					{Name: "write_file", Description: "write"},
					{Name: "edit_file", Description: "edit"},
					{Name: "search", Description: "search"},
					{Name: "execute", Description: "exec"},
					{Name: "deploy", Description: "deploy"},
				},
			},
		},
	}

	for _, ex := range examples {
		fmt.Printf("\n========== %s ==========\n", ex.name)
		result, err := complexity.Analyze(ex.req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	}

	fmt.Println("\n========== Raw JSON Input Demo ==========")
	rawJSON := `{
		"model": "claude-3-5-sonnet-20241022",
		"messages": [
			{"role": "user", "content": "hello"}
		]
	}`
	result, err := complexity.AnalyzeRaw([]byte(rawJSON))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
}

func generateLongSystemPrompt() string {
	base := "你必须按照架构规范保证兼容性，禁止违反规范，遵循不得原则，保证系统稳定性。"
	result := ""
	for i := 0; i < 3000; i++ {
		result += base
	}
	return result
}

func generateExtremeMessages() []complexity.Message {
	msgs := []complexity.Message{}
	for i := 1; i <= 5; i++ {
		msgs = append(msgs, complexity.Message{
			Role:    "user",
			Content: fmt.Sprintf("system-reminder: memory item %d", i),
		})
		msgs = append(msgs, complexity.Message{
			Role:    "assistant",
			Content: "ok",
		})
	}
	msgs = append(msgs, complexity.Message{
		Role:    "user",
		Content: "step 1. step 2. step 3. step 4. step 5. step 6. step 7. step 8. 修改 file_path: main.go file_path: handler.go file_path: router.go",
	})
	return msgs
}
