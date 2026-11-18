package challengegen

import (
	"context"
	"strings"
)

type Challenge struct {
	Description  string   `json:"description"`
	StartCode    string   `json:"starter_code"`
	TestCode     string   `json:"test_code"`
	SolutionCode string   `json:"solution_code"`
	Difficulty   int      `json:"difficulty"`
	Tags         []string `json:"tags"`
}

func (c Challenge) valid() bool {
	return strings.TrimSpace(c.Description) != "" &&
		strings.TrimSpace(c.StartCode) != "" &&
		strings.TrimSpace(c.TestCode) != "" &&
		strings.TrimSpace(c.SolutionCode) != "" &&
		c.Difficulty >= 1 && c.Difficulty <= 10 &&
		len(c.Tags) > 0
}

type Generator interface {
	GenerateChallenge(ctx context.Context, tag string) (Challenge, error)
}
