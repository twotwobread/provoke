package llm_test

import (
	"testing"

	"github.com/twotwobread/provoke/internal/config"
	"github.com/twotwobread/provoke/internal/llm"
)

func TestExtractTFContent(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "hcl block",
			input:    "Here is the config:\n```hcl\nresource \"foo\" \"bar\" {}\n```",
			expected: "resource \"foo\" \"bar\" {}",
		},
		{
			name:     "terraform block",
			input:    "```terraform\nresource \"foo\" \"bar\" {}\n```",
			expected: "resource \"foo\" \"bar\" {}",
		},
		{
			name:     "plain block",
			input:    "```\nresource \"foo\" \"bar\" {}\n```",
			expected: "resource \"foo\" \"bar\" {}",
		},
		{
			name:     "no fences",
			input:    "resource \"foo\" \"bar\" {}",
			expected: "resource \"foo\" \"bar\" {}",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := llm.ExtractTFContent(tc.input)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestNewClientUnknownProvider(t *testing.T) {
	cfg := &config.Config{}
	cfg.LLM.Provider = "unknown"
	_, err := llm.NewClient(cfg)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNewClientMissingAPIKey(t *testing.T) {
	cfg := &config.Config{}
	cfg.LLM.Provider = "claude"
	cfg.LLM.Model = "claude-sonnet-4-6"
	_, err := llm.NewClient(cfg)
	if err == nil {
		t.Fatal("expected error for missing api_key")
	}
}

func TestNewClientOllamaNoAPIKey(t *testing.T) {
	cfg := &config.Config{}
	cfg.LLM.Provider = "ollama"
	cfg.LLM.Model = "llama3.2"
	client, err := llm.NewClient(cfg)
	if err != nil {
		t.Fatalf("ollama should not require api_key, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
