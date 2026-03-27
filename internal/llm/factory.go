package llm

import (
	"fmt"

	"github.com/twotwobread/provoke/internal/config"
)

// NewClient returns an LLMClient based on the config provider.
func NewClient(cfg *config.Config) (LLMClient, error) {
	switch cfg.LLM.Provider {
	case "claude":
		return NewClaudeClient(cfg.LLM.APIKey, cfg.LLM.Model), nil
	case "openai":
		return NewOpenAIClient(cfg.LLM.APIKey, cfg.LLM.Model), nil
	case "ollama":
		return NewOllamaClient(cfg.LLM.BaseURL, cfg.LLM.Model), nil
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s (supported: claude, openai, ollama)", cfg.LLM.Provider)
	}
}
