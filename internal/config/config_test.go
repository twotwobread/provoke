package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/twotwobread/provoke/internal/config"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	content := `
llm:
  provider: claude
  model: claude-sonnet-4-6
  api_key: test-key
`
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.LLM.Provider != "claude" {
		t.Errorf("expected provider claude, got %s", cfg.LLM.Provider)
	}
	if cfg.LLM.APIKey != "test-key" {
		t.Errorf("expected api_key test-key, got %s", cfg.LLM.APIKey)
	}
}

func TestLoadOllama(t *testing.T) {
	dir := t.TempDir()
	content := `
llm:
  provider: ollama
  model: llama3.2
  base_url: http://localhost:11434
`
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.LLM.BaseURL != "http://localhost:11434" {
		t.Errorf("expected base_url, got %s", cfg.LLM.BaseURL)
	}
}
