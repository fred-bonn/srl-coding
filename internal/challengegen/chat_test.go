package challengegen

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestGenerateChallenge(t *testing.T) {
	wantChallenge := Challenge{
		Description:  "Reverse a list.",
		StartCode:    "def reverse(xs):\n    pass\n",
		TestCode:     "assert reverse([1, 2]) == [2, 1]",
		SolutionCode: "def reverse(xs):\n    return xs[::-1]",
		Difficulty:   5,
		Tags:         []string{"lists"},
	}
	challengeJSON := `{"description":"Reverse a list.","starter_code":"def reverse(xs):\n    pass\n","test_code":"assert reverse([1, 2]) == [2, 1]","solution_code":"def reverse(xs):\n    return xs[::-1]","difficulty":5,"tags":["lists"]}`

	tests := []struct {
		name    string
		body    string
		want    Challenge
		wantErr bool
	}{
		{"plain json", completionBody(challengeJSON), wantChallenge, false},
		{"fenced json", completionBody("```json\n" + challengeJSON + "\n```"), wantChallenge, false},
		{"incomplete challenge", completionBody(`{"description":"Reverse a list."}`), Challenge{}, true},
		{"api error", `{"error":{"message":"rate limited"}}`, Challenge{}, true},
		{"no choices", `{"choices":[]}`, Challenge{}, true},
		{"empty content", completionBody(""), Challenge{}, true},
		{"malformed json", "not json", Challenge{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotReq *http.Request
			var gotBody []byte
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotReq = r
				gotBody, _ = io.ReadAll(r.Body)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			g := chatGenerator{name: "test", url: srv.URL, model: "test-model", apiKey: "test-key"}
			got, err := g.GenerateChallenge(t.Context(), "loops")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("GenerateChallenge() error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateChallenge() = %+v, want %+v", got, tt.want)
			}

			if gotReq.Method != http.MethodPost {
				t.Errorf("request method = %q, want POST", gotReq.Method)
			}
			if auth := gotReq.Header.Get("Authorization"); auth != "Bearer test-key" {
				t.Errorf("Authorization header = %q, want %q", auth, "Bearer test-key")
			}
			if ct := gotReq.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type header = %q, want application/json", ct)
			}
			var req chatRequest
			if err := json.Unmarshal(gotBody, &req); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if req.Model != "test-model" {
				t.Errorf("request model = %q, want test-model", req.Model)
			}
			if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
				t.Errorf("request messages = %+v, want a single user message", req.Messages)
			}
			if !strings.Contains(req.Messages[0].Content, "loops") {
				t.Errorf("prompt does not contain the tag")
			}
		})
	}
}

func TestGenerateChallengePropagatesCanceledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	g := chatGenerator{name: "test", url: srv.URL, model: "test-model", apiKey: "test-key"}
	if _, err := g.GenerateChallenge(ctx, "loops"); err == nil {
		t.Fatal("expected an error")
	}
}

func completionBody(content string) string {
	encoded, _ := json.Marshal(content)
	return `{"choices":[{"message":{"content":` + string(encoded) + `},"finish_reason":"stop"}]}`
}
