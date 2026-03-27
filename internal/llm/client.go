package llm

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// httpClient is the shared HTTP client with a 60s timeout for all LLM providers.
var httpClient = &http.Client{Timeout: 60 * time.Second}

// LLMClient generates a Terraform file from a system prompt and user query.
type LLMClient interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// ExtractTFContent strips markdown code fences from an LLM response.
func ExtractTFContent(response string) string {
	prefixes := []string{"```hcl\n", "```terraform\n", "```\n"}
	for _, p := range prefixes {
		if idx := strings.Index(response, p); idx != -1 {
			start := idx + len(p)
			if end := strings.Index(response[start:], "```"); end != -1 {
				return strings.TrimSpace(response[start : start+end])
			}
		}
	}
	return strings.TrimSpace(response)
}
