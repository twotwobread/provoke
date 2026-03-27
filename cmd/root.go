package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	provoke_context "github.com/twotwobread/provoke/internal/context"
	"github.com/twotwobread/provoke/internal/config"
	"github.com/twotwobread/provoke/internal/llm"
	"github.com/twotwobread/provoke/internal/state"
	"github.com/twotwobread/provoke/internal/terraform"
)

var rootCmd = &cobra.Command{
	Use:   "provoke [query]",
	Short: "Provision infrastructure by invoking it in plain language",
	Args:  cobra.ArbitraryArgs,
	RunE:  runQuery,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// subcommands (init, status) register themselves via their own init()
}

func runQuery(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	query := strings.Join(args, " ")

	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config (~/.provoke/config.yaml): %w", err)
	}

	// 2. Find project dir
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectDir, err := findProjectDir(cwd)
	if err != nil {
		return err
	}

	// 3. Load state
	s, err := state.Load(projectDir)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	// 4. Read current main.tf
	tfPath := filepath.Join(projectDir, "main.tf")
	tfContent := ""
	if data, err := os.ReadFile(tfPath); err == nil {
		tfContent = string(data)
	}

	// 5. Build LLM client
	client, err := llm.NewClient(cfg)
	if err != nil {
		return err
	}

	// 6. Build prompt and call LLM
	provider := "gcp"
	if s != nil {
		provider = s.Provider
	}
	builder := provoke_context.NewBuilder(s, tfContent, provider)

	fmt.Println("Generating Terraform configuration...")
	newTF, err := client.Generate(cmd.Context(), builder.SystemPrompt(), query)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}
	// 7. Backup and write new .tf
	backupPath := tfPath + ".bak"
	hasBackup := tfContent != ""
	if hasBackup {
		if err := os.WriteFile(backupPath, []byte(tfContent), 0644); err != nil {
			return fmt.Errorf("backup main.tf: %w", err)
		}
	}
	if err := os.WriteFile(tfPath, []byte(newTF), 0644); err != nil {
		return fmt.Errorf("write main.tf: %w", err)
	}

	// 8. terraform init + plan
	runner := terraform.NewRunner(projectDir)
	fmt.Println("Running terraform init...")
	if err := runner.Init(); err != nil {
		if hasBackup {
			if rbErr := rollback(tfPath, backupPath); rbErr != nil {
				fmt.Fprintf(os.Stderr, "warning: rollback failed: %v\n", rbErr)
			}
		}
		return fmt.Errorf("terraform init failed: %w", err)
	}

	fmt.Println("Running terraform plan...")
	planOutput, err := runner.Plan()
	if err != nil {
		if hasBackup {
			if rbErr := rollback(tfPath, backupPath); rbErr != nil {
				fmt.Fprintf(os.Stderr, "warning: rollback failed: %v\n", rbErr)
			}
		}
		return fmt.Errorf("terraform plan failed: %w", err)
	}

	// 9. Show plan and confirm
	fmt.Println("\n" + summarizePlan(planOutput))
	fmt.Print("\nApply changes? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" {
		if hasBackup {
			if rbErr := rollback(tfPath, backupPath); rbErr != nil {
				fmt.Fprintf(os.Stderr, "warning: rollback failed: %v\n", rbErr)
			}
		}
		fmt.Println("Cancelled.")
		return nil
	}

	// 10. Apply with self-healing (max 2 retries)
	if err := applyWithHealing(cmd.Context(), runner, client, builder, tfPath, backupPath, hasBackup, 2); err != nil {
		return err
	}

	// 11. Update state from tfstate
	showData, err := runner.ShowJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not read terraform state: %v\n", err)
	} else if len(showData) > 0 {
		tmpPath := filepath.Join(projectDir, ".tfstate_show.json")
		if wErr := os.WriteFile(tmpPath, showData, 0644); wErr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not write state tmp file: %v\n", wErr)
		} else {
			derived, derErr := state.DeriveFromTFState(tmpPath, s)
			_ = os.Remove(tmpPath)
			if derErr != nil {
				fmt.Fprintf(os.Stderr, "warning: could not derive state: %v\n", derErr)
			} else if sErr := derived.Save(projectDir); sErr != nil {
				fmt.Fprintf(os.Stderr, "warning: could not save state: %v\n", sErr)
			}
		}
	}

	if hasBackup {
		_ = os.Remove(backupPath)
	}
	fmt.Printf("\n✓ Done. %s saved.\n", tfPath)
	return nil
}

func applyWithHealing(ctx context.Context, runner *terraform.Runner, client llm.LLMClient, builder *provoke_context.Builder, tfPath, backupPath string, hasBackup bool, maxRetries int) error {
	doRollback := func() {
		if !hasBackup {
			return
		}
		if rbErr := rollback(tfPath, backupPath); rbErr != nil {
			fmt.Fprintf(os.Stderr, "warning: rollback failed: %v\n", rbErr)
		}
	}
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := runner.Apply()
		if err == nil {
			return nil
		}
		if attempt == maxRetries {
			doRollback()
			return fmt.Errorf("terraform apply failed after %d retries: %w", maxRetries, err)
		}

		fmt.Printf("\nApply failed. Attempting self-healing (%d/%d)...\n", attempt+1, maxRetries)

		currentTF, _ := os.ReadFile(tfPath)
		fixPrompt := fmt.Sprintf(
			"The following terraform apply error occurred:\n\n%v\n\nCurrent main.tf content:\n\n%s\n\nPlease fix the terraform configuration.",
			err, string(currentTF),
		)
		fixedTF, llmErr := client.Generate(ctx, builder.SystemPrompt(), fixPrompt)
		if llmErr != nil {
			doRollback()
			return fmt.Errorf("self-healing LLM call failed: %w", llmErr)
		}
		if err := os.WriteFile(tfPath, []byte(fixedTF), 0644); err != nil {
			doRollback()
			return err
		}
	}
	return nil
}

func rollback(tfPath, backupPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}
	return os.WriteFile(tfPath, data, 0644)
}

// summarizePlan extracts key change lines from terraform plan output.
func summarizePlan(output string) string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "+") ||
			strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "~") ||
			strings.Contains(trimmed, "Plan:") {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return output
	}
	return "Changes:\n" + strings.Join(lines, "\n")
}
