package challengegen

import (
	"context"
	"strings"
)

type Challenge struct {
	Assignment   string   `json:"assignment"`
	StartCode    string   `json:"starter_code"`
	TestCode     string   `json:"test_code"`
	SolutionCode string   `json:"solution_code"`
	Difficulty   int      `json:"difficulty"`
	Topics       []string `json:"topics"`
}

func (c Challenge) valid() bool {
	return strings.TrimSpace(c.Assignment) != "" &&
		strings.TrimSpace(c.StartCode) != "" &&
		strings.TrimSpace(c.TestCode) != "" &&
		strings.TrimSpace(c.SolutionCode) != "" &&
		c.Difficulty >= 1 && c.Difficulty <= 10 &&
		len(c.Topics) > 0
}

type Generator interface {
	GenerateChallenge(ctx context.Context, topic string) (Challenge, error)
}
