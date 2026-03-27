package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Resource struct {
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Params    map[string]any `json:"params"`
	CreatedAt time.Time              `json:"created_at"`
}

type State struct {
	Project   string     `json:"project"`
	Provider  string     `json:"provider"`
	Resources []Resource `json:"resources"`
}

func stateFilePath(dir string) string {
	return filepath.Join(dir, "state.json")
}

// Load reads state.json from dir. Returns nil, nil if the file does not exist.
func Load(dir string) (*State, error) {
	data, err := os.ReadFile(stateFilePath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	return &s, nil
}

// Save writes state.json to dir.
func (s *State) Save(dir string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(stateFilePath(dir), data, 0644)
}

// tfShowOutput is the minimal shape of `terraform show -json` output.
type tfShowOutput struct {
	Values struct {
		RootModule struct {
			Resources []struct {
				Type   string                 `json:"type"`
				Name   string                 `json:"name"`
				Values map[string]any `json:"values"`
			} `json:"resources"`
		} `json:"root_module"`
	} `json:"values"`
}

// DeriveFromTFState parses a terraform show -json output file and builds a
// new State, preserving created_at from existing where available.
func DeriveFromTFState(jsonPath string, existing *State) (*State, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("read tfstate json: %w", err)
	}
	var out tfShowOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse tfstate json: %w", err)
	}

	// Build lookup map from existing state to preserve created_at.
	existingMap := map[string]time.Time{}
	if existing != nil {
		for _, r := range existing.Resources {
			existingMap[r.Type+"/"+r.Name] = r.CreatedAt
		}
	}

	derived := &State{
		Project:  existing.Project,
		Provider: existing.Provider,
	}
	for _, r := range out.Values.RootModule.Resources {
		createdAt, ok := existingMap[r.Type+"/"+r.Name]
		if !ok {
			createdAt = time.Now().UTC()
		}
		// Extract scalar params only (skip nested objects/arrays).
		params := map[string]any{}
		for k, v := range r.Values {
			switch v.(type) {
			case string, float64, bool:
				params[k] = v
			}
		}
		derived.Resources = append(derived.Resources, Resource{
			Type:      r.Type,
			Name:      r.Name,
			Params:    params,
			CreatedAt: createdAt,
		})
	}
	return derived, nil
}
