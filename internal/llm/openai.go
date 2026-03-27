package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAIClient struct {
	apiKey string
	model  string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{apiKey: apiKey, model: model}
}

func (c *OpenAIClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	body := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}
	data, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error %d: %s", resp.StatusCode, b)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response from openai")
	}
	return ExtractTFContent(result.Choices[0].Message.Content), nil
}
