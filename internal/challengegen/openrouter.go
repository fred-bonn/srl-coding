package challengegen

import (
	"errors"
	"os"
)

const (
	openRouterURL   = "https://openrouter.ai/api/v1/chat/completions"
	openRouterModel = "nvidia/nemotron-3.5-super-vl:free"
)

func NewOpenRouterGenerator() (Generator, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENROUTER_API_KEY is not set")
	}
	return chatGenerator{
		name:   "OpenRouter",
		url:    openRouterURL,
		model:  openRouterModel,
		apiKey: apiKey,
	}, nil
}
