package context

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/twotwobread/provoke/internal/state"
)

// Builder assembles the system prompt sent to the LLM.
type Builder struct {
	state    *state.State
	tfFile   string
	provider string
}

func NewBuilder(s *state.State, tfFile, provider string) *Builder {
	return &Builder{state: s, tfFile: tfFile, provider: provider}
}

// SystemPrompt builds the system prompt with current state and .tf content.
func (b *Builder) SystemPrompt() string {
	stateStr := "no resources deployed yet"
	if b.state != nil && len(b.state.Resources) > 0 {
		data, _ := json.MarshalIndent(b.state.Resources, "", "  ")
		stateStr = string(data)
	}

	tfStr := "none"
	if b.tfFile != "" {
		tfStr = b.tfFile
	}

	return fmt.Sprintf(`You are a Terraform expert.
Current cloud provider: %s
Current date: %s

Currently deployed resources:
%s

Current main.tf:
%s

Instructions:
- Return ONLY a valid, complete Terraform HCL file. Do not include explanations.
- Wrap the output in a markdown hcl code block.
- Preserve existing resources unless the user explicitly asks to remove them.
- Use the cheapest/free tier options unless the user specifies otherwise.`,
		b.provider,
		time.Now().UTC().Format("2006-01-02"),
		stateStr,
		tfStr,
	)
}
