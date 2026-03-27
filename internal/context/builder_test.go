package context_test

import (
	"strings"
	"testing"
	"time"

	provoke_context "github.com/twotwobread/provoke/internal/context"
	"github.com/twotwobread/provoke/internal/state"
)

func TestSystemPromptContainsState(t *testing.T) {
	s := &state.State{
		Project:  "my-app",
		Provider: "gcp",
		Resources: []state.Resource{
			{
				Type:      "google_container_cluster",
				Name:      "main",
				Params:    map[string]any{"node_count": float64(3)},
				CreatedAt: time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	b := provoke_context.NewBuilder(s, "resource \"google_container_cluster\" \"main\" {}", "gcp")
	prompt := b.SystemPrompt()

	if !strings.Contains(prompt, "google_container_cluster") {
		t.Error("prompt should contain resource type")
	}
	if !strings.Contains(prompt, "gcp") {
		t.Error("prompt should contain provider")
	}
	if !strings.Contains(prompt, "2026-03-27") {
		t.Error("prompt should contain date")
	}
}

func TestSystemPromptEmptyState(t *testing.T) {
	b := provoke_context.NewBuilder(nil, "", "gcp")
	prompt := b.SystemPrompt()

	if !strings.Contains(prompt, "Terraform expert") {
		t.Error("prompt should contain Terraform expert")
	}
	if !strings.Contains(prompt, "no resources") {
		t.Error("prompt should indicate no resources when state is nil")
	}
}
