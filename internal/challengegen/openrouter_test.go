package challengegen

import "testing"

func TestNewOpenRouterGenerator(t *testing.T) {
	t.Run("missing api key", func(t *testing.T) {
		t.Setenv("OPENROUTER_API_KEY", "")
		if _, err := NewOpenRouterGenerator(); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("configured from environment", func(t *testing.T) {
		t.Setenv("OPENROUTER_API_KEY", "test-key")
		g, err := NewOpenRouterGenerator()
		if err != nil {
			t.Fatalf("NewOpenRouterGenerator() error: %v", err)
		}
		gg, ok := g.(chatGenerator)
		if !ok {
			t.Fatalf("NewOpenRouterGenerator() = %T, want chatGenerator", g)
		}
		if gg.name != "OpenRouter" || gg.url != openRouterURL || gg.model != openRouterModel || gg.apiKey != "test-key" {
			t.Errorf("generator = name %q, url %q, model %q, apiKey %q; want name OpenRouter, url %q, model %q, apiKey test-key",
				gg.name, gg.url, gg.model, gg.apiKey, openRouterURL, openRouterModel)
		}
	})
}
