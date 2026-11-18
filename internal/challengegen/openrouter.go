package challengegen

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

//go:embed prompt.md
var promptTemplate string

//go:embed samples/*.json
var samplesFS embed.FS

const (
	openRouterURL   = "https://openrouter.ai/api/v1/chat/completions"
	openRouterModel = "nvidia/nemotron-3.5-super-vl:free"
)

type openRouterGenerator struct {
	apiKey string
}

func NewOpenRouterGenerator() (Generator, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENROUTER_API_KEY is not set")
	}
	return &openRouterGenerator{apiKey: apiKey}, nil
}

func (g *openRouterGenerator) GenerateChallenge(ctx context.Context, tag string) (Challenge, error) {
	prompt, err := buildPrompt(tag)
	if err != nil {
		return Challenge{}, err
	}

	body, err := json.Marshal(chatRequest{
		Model:       openRouterModel,
		Messages:    []chatMessage{{Role: "user", Content: prompt}},
		Temperature: 0.2,
		MaxTokens:   4000,
	})
	if err != nil {
		return Challenge{}, fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(body))
	if err != nil {
		return Challenge{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Challenge{}, fmt.Errorf("calling OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	content, err := decodeResponse(resp)
	if err != nil {
		return Challenge{}, err
	}

	var challenge Challenge
	if err := json.Unmarshal([]byte(stripJSONFences(content)), &challenge); err != nil {
		return Challenge{}, fmt.Errorf("parsing generated challenge: %w", err)
	}
	if !challenge.valid() {
		return Challenge{}, errors.New("generated challenge is incomplete: every field must be non-empty and tags must have at least one entry")
	}
	return challenge, nil
}

func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[i+1:]
	}
	return strings.TrimSuffix(strings.TrimSpace(s), "```")
}

func (c Challenge) valid() bool {
	return strings.TrimSpace(c.Description) != "" &&
		strings.TrimSpace(c.StartCode) != "" &&
		strings.TrimSpace(c.TestCode) != "" &&
		strings.TrimSpace(c.SolutionCode) != "" &&
		len(c.Tags) > 0
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func decodeResponse(resp *http.Response) (string, error) {
	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("decoding OpenRouter response (status %d): %w", resp.StatusCode, err)
	}
	if cr.Error != nil && cr.Error.Message != "" {
		return "", fmt.Errorf("OpenRouter error: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 || cr.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("OpenRouter returned no content (status %d)", resp.StatusCode)
	}
	return cr.Choices[0].Message.Content, nil
}
