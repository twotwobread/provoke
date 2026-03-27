package terraform

import (
	"fmt"
	"os/exec"
)

// Executor abstracts os/exec to allow testing without running real terraform.
type Executor interface {
	Run(dir string, args ...string) ([]byte, error)
}

type osExecutor struct{}

func (e *osExecutor) Run(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// Runner executes terraform commands in a given project directory.
type Runner struct {
	dir      string
	executor Executor
}

func NewRunner(dir string) *Runner {
	return &Runner{dir: dir, executor: &osExecutor{}}
}

// NewRunnerWithExecutor creates a Runner with a custom executor (for testing).
func NewRunnerWithExecutor(dir string, exec Executor) *Runner {
	return &Runner{dir: dir, executor: exec}
}

// Plan runs `terraform plan` and returns the output.
func (r *Runner) Plan() (string, error) {
	out, err := r.executor.Run(r.dir, "plan", "-no-color")
	if err != nil {
		return "", fmt.Errorf("terraform plan: %w\n%s", err, out)
	}
	return string(out), nil
}

// Apply runs `terraform apply -auto-approve`.
func (r *Runner) Apply() error {
	out, err := r.executor.Run(r.dir, "apply", "-auto-approve", "-no-color")
	if err != nil {
		return fmt.Errorf("terraform apply: %w\n%s", err, out)
	}
	return nil
}

// ShowJSON runs `terraform show -json` and returns raw output.
func (r *Runner) ShowJSON() ([]byte, error) {
	out, err := r.executor.Run(r.dir, "show", "-json")
	if err != nil {
		return nil, fmt.Errorf("terraform show: %w\n%s", err, out)
	}
	return out, nil
}

// Init runs `terraform init`.
func (r *Runner) Init() error {
	out, err := r.executor.Run(r.dir, "init", "-no-color")
	if err != nil {
		return fmt.Errorf("terraform init: %w\n%s", err, out)
	}
	return nil
}
