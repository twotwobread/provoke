package llm_test

import (
	"testing"

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
