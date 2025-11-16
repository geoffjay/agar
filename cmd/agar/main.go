package main

import (
	"github.com/geoffjay/agar/cmd/agar/cmd"
	"github.com/geoffjay/agar/cmd/agar/internal/commands"
)

func main() {
	// Register subcommands
	cmd.AddCommand(commands.InitCmd())

	// Execute root command
	cmd.Execute()
}
