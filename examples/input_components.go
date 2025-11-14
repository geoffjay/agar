package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/geoffjay/agar/tui"
)

func main() {
	// Example 1: Yes/No Input
	fmt.Println("=== Example 1: Yes/No Input ===\n")
	yesNoModel := tui.NewYesNoInput(
		"Do you want to continue?",
		"This will proceed with the next step",
	)
	p := tea.NewProgram(yesNoModel)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	yesNoResult := finalModel.(tui.YesNoModel)
	fmt.Printf("\nYou answered: %s\n\n", yesNoResult.GetAnswerString())

	// Example 2: Single-line Text Input
	fmt.Println("=== Example 2: Single-line Text ===\n")
	textModel := tui.NewTextInput(
		"What is your name?",
		"Enter your full name",
		tui.SingleLine,
	)
	p = tea.NewProgram(textModel)
	finalModel, err = p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	textResult := finalModel.(tui.TextModel)
	fmt.Printf("\nYou entered: %s\n\n", textResult.GetAnswer())

	// Example 3: Multi-line Text Input
	fmt.Println("=== Example 3: Multi-line Text ===\n")
	multiTextModel := tui.NewTextInput(
		"Describe your project:",
		"Provide a detailed description (Ctrl+D when done)",
		tui.MultiLine,
	)
	p = tea.NewProgram(multiTextModel)
	finalModel, err = p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	multiTextResult := finalModel.(tui.TextModel)
	fmt.Printf("\nYou entered:\n%s\n\n", multiTextResult.GetAnswer())

	// Example 4: Options Input
	fmt.Println("=== Example 4: Options Selection ===\n")
	optionsModel := tui.NewOptionsInput(
		"What is your favorite programming language?",
		"Select one option from the list",
		[]string{
			"Go",
			"Rust",
			"Python",
			"TypeScript",
			"Other",
		},
	)
	p = tea.NewProgram(optionsModel)
	finalModel, err = p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	optionsResult := finalModel.(tui.OptionsModel)
	fmt.Printf("\nYou selected: %s (index: %d)\n\n",
		optionsResult.GetAnswer(),
		optionsResult.GetSelectedIndex())

	fmt.Println("=== All examples completed! ===")
}
