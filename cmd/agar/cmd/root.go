package cmd

import (
	"fmt"
	"os"

	"github.com/geoffjay/agar/cmd/agar/internal/app"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agar",
	Short: "A framework for building AI agent applications",
	Long: `Agar is a comprehensive framework for building AI agent applications
with TUI components and tool management.

Running 'agar' without arguments launches the interactive TUI interface.
Use subcommands for specific operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Launch TUI when no subcommand is specified
		if err := app.RunTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.agar.yaml)")
}
