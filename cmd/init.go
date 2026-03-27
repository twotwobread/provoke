package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a provoke project in the current directory",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Project name: default to current directory name
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defaultName := filepath.Base(cwd)

	fmt.Printf("Project name [%s]: ", defaultName)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}

	// Cloud provider
	fmt.Print("Cloud provider (gcp/aws/azure) [gcp]: ")
	provider, _ := reader.ReadString('\n')
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "gcp"
	}

	// Create .provoke/<project-name>/ directory
	projectDir := filepath.Join(cwd, ".provoke", name)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("create project dir: %w", err)
	}

	// Write initial state.json
	initialState := map[string]any{
		"project":   name,
		"provider":  provider,
		"resources": []any{},
	}
	stateData, _ := json.MarshalIndent(initialState, "", "  ")
	statePath := filepath.Join(projectDir, "state.json")
	if err := os.WriteFile(statePath, stateData, 0644); err != nil {
		return fmt.Errorf("write state.json: %w", err)
	}

	// Create empty main.tf and variables.tf
	for _, fname := range []string{"main.tf", "variables.tf"} {
		path := filepath.Join(projectDir, fname)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(""), 0644); err != nil {
				return fmt.Errorf("create %s: %w", fname, err)
			}
		}
	}

	fmt.Printf("\n✓ Initialized provoke project '%s' (provider: %s)\n", name, provider)
	fmt.Printf("  Project dir: .provoke/%s/\n", name)
	fmt.Println("\nNext: provoke \"<your infrastructure request>\"")
	return nil
}
