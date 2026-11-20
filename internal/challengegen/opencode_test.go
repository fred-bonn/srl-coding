package challengegen

import "testing"

func TestNewOpencodeGenerator(t *testing.T) {
	t.Run("missing api key", func(t *testing.T) {
		t.Setenv("OPENCODE_API_KEY", "")
		if _, err := NewOpencodeGenerator(); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("configured from environment", func(t *testing.T) {
		t.Setenv("OPENCODE_API_KEY", "test-key")
		g, err := NewOpencodeGenerator()
		if err != nil {
			t.Fatalf("NewOpencodeGenerator() error: %v", err)
		}
		gg, ok := g.(chatGenerator)
		if !ok {
			t.Fatalf("NewOpencodeGenerator() = %T, want chatGenerator", g)
		}
		if gg.name != "OpenCode" || gg.url != opencodeURL || gg.model != opencodeModel || gg.apiKey != "test-key" {
			t.Errorf("generator = name %q, url %q, model %q, apiKey %q; want name OpenCode, url %q, model %q, apiKey test-key",
				gg.name, gg.url, gg.model, gg.apiKey, opencodeURL, opencodeModel)
		}
	})
}
