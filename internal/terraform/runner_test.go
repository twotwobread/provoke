package terraform_test

import (
	"testing"

	"github.com/twotwobread/provoke/internal/terraform"
)

type mockExecutor struct {
	outputs map[string][]byte
	errors  map[string]error
}

func (m *mockExecutor) Run(dir string, args ...string) ([]byte, error) {
	key := args[0]
	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	if out, ok := m.outputs[key]; ok {
		return out, nil
	}
	return []byte{}, nil
}

func TestPlanSuccess(t *testing.T) {
	exec := &mockExecutor{
		outputs: map[string][]byte{
			"plan": []byte("Plan: 1 to add, 0 to change, 0 to destroy."),
		},
	}
	runner := terraform.NewRunnerWithExecutor("/tmp/project", exec)
	output, err := runner.Plan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty plan output")
	}
}

func TestApplySuccess(t *testing.T) {
	exec := &mockExecutor{
		outputs: map[string][]byte{
			"apply": []byte("Apply complete!"),
		},
	}
	runner := terraform.NewRunnerWithExecutor("/tmp/project", exec)
	if err := runner.Apply(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShowJSON(t *testing.T) {
	expected := []byte(`{"format_version":"1.0"}`)
	exec := &mockExecutor{
		outputs: map[string][]byte{"show": expected},
	}
	runner := terraform.NewRunnerWithExecutor("/tmp/project", exec)
	out, err := runner.ShowJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != string(expected) {
		t.Errorf("expected %s, got %s", expected, out)
	}
}
