package main

import (
	"os"

	"github.com/geoffjay/agar/cmd/agar/cmd"
	"github.com/geoffjay/agar/cmd/agar/internal/commands"
)

func main() {
	os.Setenv("BAML_LOG", "warn")

	// Register subcommands
	cmd.AddCommand(commands.InitCmd())

	// Execute root command
	cmd.Execute()
}
