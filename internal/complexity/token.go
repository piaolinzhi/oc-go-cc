package complexity

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

type TokenCounter struct {
	encoder *tiktoken.Tiktoken
}

func NewTokenCounter() (*TokenCounter, error) {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to get cl100k_base encoding: %w", err)
	}
	return &TokenCounter{encoder: enc}, nil
}

func (tc *TokenCounter) Count(text string) int {
	return len(tc.encoder.Encode(text, nil, nil))
}

func (tc *TokenCounter) CountMessages(systemTexts []string, messages []Message) int {
	total := 3
	for _, s := range systemTexts {
		total += tc.Count(s) + 5
	}
	for _, msg := range messages {
		total += tc.Count(msg.Content) + 5
	}
	return total
}
