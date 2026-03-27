package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ClaudeClient struct {
	apiKey string
	model  string
}

func NewClaudeClient(apiKey, model string) *ClaudeClient {
	return &ClaudeClient{apiKey: apiKey, model: model}
}

func (c *ClaudeClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	body := map[string]any{
		"model":      c.model,
		"max_tokens": 4096,
		"system":     systemPrompt,
		"messages":   []map[string]string{{"role": "user", "content": userPrompt}},
	}
	data, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("claude error %d: %s", resp.StatusCode, b)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode claude response: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from claude")
	}
	return ExtractTFContent(result.Content[0].Text), nil
}
