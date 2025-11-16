package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agar",
	Short: "A CLI utility for creating new AI tools",
	Long: `agar is a CLI utility for simplifying the creation of agar
	applications. It provides commands for initializing new applications that use
	the agar framework.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior when no subcommand is specified
		cmd.Help()
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
