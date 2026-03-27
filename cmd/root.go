package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "provoke [query]",
	Short: "Provision infrastructure by invoking it in plain language",
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
		// Join all args so both quoted and unquoted queries work:
		// provoke "scale to 2 nodes"  →  args[0]
		// provoke scale to 2 nodes    →  strings.Join(args, " ")
		fmt.Println("query:", strings.Join(args, " ")) // placeholder — implemented in Task 9
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
