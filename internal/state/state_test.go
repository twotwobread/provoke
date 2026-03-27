package state_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/twotwobread/provoke/internal/state"
)

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()

	s := &state.State{
		Project:  "test-project",
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

	if err := s.Save(dir); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := state.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.Project != "test-project" {
		t.Errorf("expected project test-project, got %s", loaded.Project)
	}
	if len(loaded.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(loaded.Resources))
	}
	if loaded.Resources[0].Type != "google_container_cluster" {
		t.Errorf("unexpected resource type: %s", loaded.Resources[0].Type)
	}
}

func TestLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	s, err := state.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != nil {
		t.Errorf("expected nil for missing state, got %+v", s)
	}
}

func TestDeriveFromTFState(t *testing.T) {
	tfstate := map[string]any{
		"format_version": "1.0",
		"values": map[string]any{
			"root_module": map[string]any{
				"resources": []any{
					map[string]any{
						"type": "google_container_cluster",
						"name": "main",
						"values": map[string]any{
							"name":     "main-cluster",
							"location": "us-central1",
						},
					},
				},
			},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "terraform.tfstate.json")
	data, _ := json.Marshal(tfstate)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	existing := &state.State{Project: "my-app", Provider: "gcp"}
	derived, err := state.DeriveFromTFState(path, existing)
	if err != nil {
		t.Fatalf("derive: %v", err)
	}

	if len(derived.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(derived.Resources))
	}
	if derived.Resources[0].Type != "google_container_cluster" {
		t.Errorf("unexpected type: %s", derived.Resources[0].Type)
	}
	if derived.Resources[0].CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}
