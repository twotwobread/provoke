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
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectDir, err := findProjectDir(cwd)
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

// findProjectDir finds the single .provoke/<project>/ directory under root.
// Accepts root as a parameter to allow testing without os.Chdir.
func findProjectDir(root string) (string, error) {
	provokeDir := filepath.Join(root, ".provoke")
	entries, err := os.ReadDir(provokeDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("no .provoke/ directory found. Run 'provoke init' first")
	}
	if err != nil {
		return "", err
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(provokeDir, e.Name()))
		}
	}
	if len(dirs) == 0 {
		return "", fmt.Errorf("no project found in .provoke/. Run 'provoke init' first")
	}
	if len(dirs) > 1 {
		names := make([]string, len(dirs))
		for i, d := range dirs {
			names[i] = filepath.Base(d)
		}
		return "", fmt.Errorf("multiple projects found in .provoke/: %v — run from the correct project root", names)
	}
	return dirs[0], nil
}
