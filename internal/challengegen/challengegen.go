package challengegen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Challenge struct {
	Description  string   `json:"description"`
	StartCode    string   `json:"starter_code"`
	TestCode     string   `json:"test_code"`
	SolutionCode string   `json:"solution_code"`
	Tags         []string `json:"tags"`
}

func (c Challenge) Format() (string, error) {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", fmt.Errorf("formatting challenge: %w", err)
	}
	return string(b), nil
}

type Generator interface {
	GenerateChallenge(ctx context.Context, tag string) (Challenge, error)
}

func buildPrompt(tag string) (string, error) {
	_, template, ok := strings.Cut(promptTemplate, "\n---\n")
	if !ok {
		return "", errors.New("prompt.md is missing the --- separator before the template")
	}

	var b strings.Builder
	b.WriteString(strings.Replace(template, "{TOPIC}", tag, 1))
	b.WriteString("\n\nExample outputs:\n")

	entries, err := samplesFS.ReadDir("samples")
	if err != nil {
		return "", fmt.Errorf("reading samples: %w", err)
	}
	for _, entry := range entries {
		sample, err := samplesFS.ReadFile("samples/" + entry.Name())
		if err != nil {
			return "", fmt.Errorf("reading sample %s: %w", entry.Name(), err)
		}
		b.WriteString("\n")
		b.Write(sample)
		b.WriteString("\n")
	}
	return b.String(), nil
}
