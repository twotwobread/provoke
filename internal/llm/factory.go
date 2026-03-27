package llm

import (
	"fmt"

	"github.com/twotwobread/provoke/internal/config"
)

// NewClient returns an LLMClient based on the config provider.
func NewClient(cfg *config.Config) (LLMClient, error) {
	switch cfg.LLM.Provider {
	case "claude":
		if cfg.LLM.APIKey == "" {
			return nil, fmt.Errorf("claude provider requires api_key in config")
		}
		if cfg.LLM.Model == "" {
			return nil, fmt.Errorf("claude provider requires model in config")
		}
		return NewClaudeClient(cfg.LLM.APIKey, cfg.LLM.Model), nil
	case "openai":
		if cfg.LLM.APIKey == "" {
			return nil, fmt.Errorf("openai provider requires api_key in config")
		}
		if cfg.LLM.Model == "" {
			return nil, fmt.Errorf("openai provider requires model in config")
		}
		return NewOpenAIClient(cfg.LLM.APIKey, cfg.LLM.Model), nil
	case "ollama":
		if cfg.LLM.Model == "" {
			return nil, fmt.Errorf("ollama provider requires model in config")
		}
		return NewOllamaClient(cfg.LLM.BaseURL, cfg.LLM.Model), nil
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s (supported: claude, openai, ollama)", cfg.LLM.Provider)
	}
}
