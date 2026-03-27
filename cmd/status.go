package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/twotwobread/provoke/internal/state"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current deployed resources",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	projectDir, err := findProjectDir()
	if err != nil {
		return err
	}

	s, err := state.Load(projectDir)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	if s == nil || len(s.Resources) == 0 {
		fmt.Println("No resources deployed.")
		return nil
	}

	fmt.Printf("Project: %s (provider: %s)\n\n", s.Project, s.Provider)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TYPE\tNAME\tCREATED")
	for _, r := range s.Resources {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			r.Type, r.Name, r.CreatedAt.Format("2006-01-02 15:04"))
	}
	w.Flush()
	return nil
}

// findProjectDir finds the first .provoke/<project>/ directory in cwd.
func findProjectDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	provokDir := filepath.Join(cwd, ".provoke")
	entries, err := os.ReadDir(provokDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("no .provoke/ directory found. Run 'provoke init' first")
	}
	if err != nil {
		return "", err
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(provokDir, e.Name()))
		}
	}
	if len(dirs) == 0 {
		return "", fmt.Errorf("no project found in .provoke/. Run 'provoke init' first")
	}
	if len(dirs) > 1 {
		return "", fmt.Errorf("multiple projects found. Use --project flag to specify one")
	}
	return dirs[0], nil
}
