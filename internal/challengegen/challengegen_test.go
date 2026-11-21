package challengegen

import "testing"

func TestChallengeValid(t *testing.T) {
	valid := Challenge{
		Assignment:   "Reverse a list.",
		StartCode:    "def reverse(xs):\n    pass\n",
		TestCode:     "assert reverse([1, 2]) == [2, 1]\n",
		SolutionCode: "def reverse(xs):\n    return xs[::-1]\n",
		Difficulty:   5,
		Topics:       []string{"lists"},
	}

	tests := []struct {
		name   string
		mutate func(*Challenge)
		want   bool
	}{
		{"valid", nil, true},
		{"difficulty lower bound", func(c *Challenge) { c.Difficulty = 1 }, true},
		{"difficulty upper bound", func(c *Challenge) { c.Difficulty = 10 }, true},
		{"difficulty zero", func(c *Challenge) { c.Difficulty = 0 }, false},
		{"difficulty above ten", func(c *Challenge) { c.Difficulty = 11 }, false},
		{"blank assignment", func(c *Challenge) { c.Assignment = "  \n\t" }, false},
		{"blank start code", func(c *Challenge) { c.StartCode = "   " }, false},
		{"blank test code", func(c *Challenge) { c.TestCode = "   " }, false},
		{"blank solution code", func(c *Challenge) { c.SolutionCode = "" }, false},
		{"no topics", func(c *Challenge) { c.Topics = nil }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := valid
			if tt.mutate != nil {
				tt.mutate(&c)
			}
			if got := c.valid(); got != tt.want {
				t.Errorf("valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
