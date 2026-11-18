package challengegen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type chatGenerator struct {
	name   string
	url    string
	model  string
	apiKey string
}

func (g chatGenerator) GenerateChallenge(ctx context.Context, tag string) (Challenge, error) {
	prompt, err := buildPrompt(tag)
	if err != nil {
		return Challenge{}, err
	}

	body, err := json.Marshal(chatRequest{
		Model:       g.model,
		Messages:    []chatMessage{{Role: "user", Content: prompt}},
		Temperature: 0.2,
		MaxTokens:   16000,
	})
	if err != nil {
		return Challenge{}, fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.url, bytes.NewReader(body))
	if err != nil {
		return Challenge{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Challenge{}, fmt.Errorf("calling %s: %w", g.name, err)
	}
	defer resp.Body.Close()

	content, err := decodeResponse(g.name, resp)
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
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func decodeResponse(name string, resp *http.Response) (string, error) {
	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("decoding %s response (status %d): %w", name, resp.StatusCode, err)
	}
	if cr.Error != nil && cr.Error.Message != "" {
		return "", fmt.Errorf("%s error: %s", name, cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("%s returned no choices (status %d)", name, resp.StatusCode)
	}
	choice := cr.Choices[0]
	if choice.Message.Content == "" {
		if choice.FinishReason == "length" {
			return "", fmt.Errorf("%s hit the max token limit before producing content (status %d)", name, resp.StatusCode)
		}
		return "", fmt.Errorf("%s returned no content (status %d, finish reason %q)", name, resp.StatusCode, choice.FinishReason)
	}
	return choice.Message.Content, nil
}
