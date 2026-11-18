package challengegen

import (
	"errors"
	"os"
)

const (
	opencodeURL   = "https://opencode.ai/inference/openai/v1/chat/completions"
	opencodeModel = "deepseek-v4.1-flash"
)

func NewOpencodeGenerator() (Generator, error) {
	apiKey := os.Getenv("OPENCODE_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENCODE_API_KEY is not set")
	}
	return chatGenerator{
		name:   "OpenCode",
		url:    opencodeURL,
		model:  opencodeModel,
		apiKey: apiKey,
	}, nil
}
