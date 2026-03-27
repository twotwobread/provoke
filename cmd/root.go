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
	newTF, err := client.Generate(context.Background(), builder.SystemPrompt(), query)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}
	newTF = llm.ExtractTFContent(newTF)

	// 7. Backup and write new .tf
	backupPath := tfPath + ".bak"
	_ = os.WriteFile(backupPath, []byte(tfContent), 0644)
	if err := os.WriteFile(tfPath, []byte(newTF), 0644); err != nil {
		return fmt.Errorf("write main.tf: %w", err)
	}

	// 8. terraform init + plan
	runner := terraform.NewRunner(projectDir)
	fmt.Println("Running terraform init...")
	if err := runner.Init(); err != nil {
		_ = rollback(tfPath, backupPath)
		return fmt.Errorf("terraform init failed: %w", err)
	}

	fmt.Println("Running terraform plan...")
	planOutput, err := runner.Plan()
	if err != nil {
		_ = rollback(tfPath, backupPath)
		return fmt.Errorf("terraform plan failed: %w", err)
	}

	// 9. Show plan and confirm
	fmt.Println("\n" + summarizePlan(planOutput))
	fmt.Print("\nApply changes? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" {
		_ = rollback(tfPath, backupPath)
		fmt.Println("Cancelled.")
		return nil
	}

	// 10. Apply with self-healing (max 2 retries)
	if err := applyWithHealing(runner, client, builder, tfPath, backupPath, 2); err != nil {
		return err
	}

	// 11. Update state from tfstate
	showData, err := runner.ShowJSON()
	if err == nil && len(showData) > 0 {
		tmpPath := filepath.Join(projectDir, ".tfstate_show.json")
		_ = os.WriteFile(tmpPath, showData, 0644)
		derived, derErr := state.DeriveFromTFState(tmpPath, s)
		_ = os.Remove(tmpPath)
		if derErr == nil {
			_ = derived.Save(projectDir)
		}
	}

	_ = os.Remove(backupPath)
	fmt.Printf("\n✓ Done. %s saved.\n", tfPath)
	return nil
}

func applyWithHealing(runner *terraform.Runner, client llm.LLMClient, builder *provoke_context.Builder, tfPath, backupPath string, maxRetries int) error {
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := runner.Apply()
		if err == nil {
			return nil
		}
		if attempt == maxRetries {
			_ = rollback(tfPath, backupPath)
			return fmt.Errorf("terraform apply failed after %d retries: %w", maxRetries, err)
		}

		fmt.Printf("\nApply failed. Attempting self-healing (%d/%d)...\n", attempt+1, maxRetries)

		fixPrompt := fmt.Sprintf("The following terraform apply error occurred:\n\n%v\n\nPlease fix the terraform configuration.", err)
		fixedTF, llmErr := client.Generate(context.Background(), builder.SystemPrompt(), fixPrompt)
		if llmErr != nil {
			_ = rollback(tfPath, backupPath)
			return fmt.Errorf("self-healing LLM call failed: %w", llmErr)
		}
		fixedTF = llm.ExtractTFContent(fixedTF)
		if err := os.WriteFile(tfPath, []byte(fixedTF), 0644); err != nil {
			_ = rollback(tfPath, backupPath)
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
