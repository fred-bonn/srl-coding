package challengegen

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestStripJSONFences(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no fences", `{"difficulty": 3}`, `{"difficulty": 3}`},
		// A closed fence leaves the newline that preceded it.
		{"json fence", "```json\n{\"difficulty\": 3}\n```", "{\"difficulty\": 3}\n"},
		{"plain fence", "```\n{\"difficulty\": 3}\n```", "{\"difficulty\": 3}\n"},
		{"fence with surrounding whitespace", "  ```json\n{\"difficulty\": 3}\n```  \n", "{\"difficulty\": 3}\n"},
		{"unclosed fence", "```json\n{\"difficulty\": 3}", `{"difficulty": 3}`},
		{"fence only", "```", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripJSONFences(tt.in); got != tt.want {
				t.Errorf("stripJSONFences() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name string
		tag  string
	}{
		{"single word", "loops"},
		{"two words", "list comprehensions"},
		{"punctuation", "don't-care"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := buildPrompt(tt.tag)
			if err != nil {
				t.Fatalf("buildPrompt() error: %v", err)
			}
			if !strings.Contains(prompt, tt.tag) {
				t.Errorf("prompt does not contain the tag %q", tt.tag)
			}
			if strings.Contains(prompt, "{TOPIC}") {
				t.Errorf("prompt still contains the {TOPIC} placeholder")
			}
			if !strings.Contains(prompt, "Example outputs:") {
				t.Errorf("prompt is missing the example outputs section")
			}
			entries, err := samplesFS.ReadDir("samples")
			if err != nil {
				t.Fatalf("reading samples: %v", err)
			}
			for _, entry := range entries {
				sample, err := samplesFS.ReadFile("samples/" + entry.Name())
				if err != nil {
					t.Fatalf("reading sample %s: %v", entry.Name(), err)
				}
				if !strings.Contains(prompt, string(sample)) {
					t.Errorf("prompt is missing sample %s", entry.Name())
				}
			}
		})
	}
}

func TestDecodeResponse(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr bool
	}{
		{"content", `{"choices":[{"message":{"content":"hello"},"finish_reason":"stop"}]}`, "hello", false},
		{"api error", `{"error":{"message":"rate limited"}}`, "", true},
		{"no choices", `{"choices":[]}`, "", true},
		{"empty content at max tokens", `{"choices":[{"message":{"content":""},"finish_reason":"length"}]}`, "", true},
		{"empty content", `{"choices":[{"message":{"content":""},"finish_reason":"stop"}]}`, "", true},
		{"malformed json", "not json", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}
			got, err := decodeResponse("test", resp)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeResponse() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("decodeResponse() = %q, want %q", got, tt.want)
			}
		})
	}
}
