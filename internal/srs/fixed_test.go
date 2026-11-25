package srs

import (
	"testing"
	"time"
)

func TestFixedNextDue(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	day := 24 * time.Hour

	tests := []struct {
		name    string
		correct []bool
		want    time.Duration
	}{
		{"no reviews yet", nil, 0},
		{"failed intro", []bool{false}, day},
		{"first success", []bool{true}, day},
		{"two successes", []bool{true, true}, 3 * day},
		{"three successes", []bool{true, true, true}, 7 * day},
		{"four successes", []bool{true, true, true, true}, 14 * day},
		{"five successes", []bool{true, true, true, true, true}, 30 * day},
		{"successes cap out", []bool{true, true, true, true, true, true, true}, 30 * day},
		{"failure resets streak", []bool{true, true, true, false, true, true}, 3 * day},
		{"failure after cap", []bool{true, true, true, true, true, true, false}, day},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviews := make([]Review, len(tt.correct))
			for i, correct := range tt.correct {
				reviews[i] = Review{Correct: correct, ReviewedAt: now.Add(-time.Duration(len(tt.correct)-i) * day)}
			}
			got := Fixed{}.NextDue(reviews, now)
			if !got.Equal(now.Add(tt.want)) {
				t.Errorf("NextDue = %v, want %v", got, now.Add(tt.want))
			}
		})
	}
}
