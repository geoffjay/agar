# TUI Input Components

Reusable input components for building interactive CLI applications with bubbletea.

## Overview

This package provides ready-to-use input components that handle user input with a consistent, polished interface. All components follow the `tea.Model` interface and can be used standalone or composed into larger applications.

## Components

### 1. Yes/No Input (`YesNoModel`)

A simple yes/no choice component with keyboard shortcuts.

```go
model := tui.NewYesNoInput(
    "Do you want to continue?",
    "This will proceed with the next step",
)

p := tea.NewProgram(model)
finalModel, _ := p.Run()
result := finalModel.(tui.YesNoModel)

answer := result.GetAnswer()        // bool
answerStr := result.GetAnswerString() // "yes" or "no"
```

**Features:**

- Arrow keys (↑/↓) or vim keys (j/k) to navigate
- Press `y` for quick "yes", `n` for quick "no"
- Enter or Space to confirm selection
- Checkbox-style visual indicator

**Visual Example:**

```
Do you want to continue?
This will proceed with the next step

───────────────────────────────────

[x] Yes
[ ] No

───────────────────────────────────

↑/↓ or h/j/k/l to select • Enter/Space to confirm • y/n for quick answer • Esc to cancel
```

### 2. Text Input (`TextModel`)

Text input component supporting both single-line and multi-line input.

#### Single-line Mode

```go
model := tui.NewTextInput(
    "What is your name?",
    "Enter your full name",
    tui.SingleLine,
)

p := tea.NewProgram(model)
finalModel, _ := p.Run()
result := finalModel.(tui.TextModel)

name := result.GetAnswer() // string
```

**Features:**

- Type text directly
- Backspace to delete
- Enter to submit
- Automatic text wrapping at terminal width

**Visual Example:**

```
What is your name?
Enter your full name

───────────────────────────────────

> John Doe█

───────────────────────────────────

Enter to submit • Esc to cancel
```

#### Multi-line Mode

```go
model := tui.NewTextInput(
    "Describe your project:",
    "Provide a detailed description",
    tui.MultiLine,
)

p := tea.NewProgram(model)
finalModel, _ := p.Run()
result := finalModel.(tui.TextModel)

description := result.GetAnswer() // string (may contain \n)
```

**Features:**

- Enter creates new line
- Ctrl+D to submit
- Automatic text wrapping at terminal width
- Preserves user's line breaks

**Visual Example:**

```
Describe your project:
Provide a detailed description

───────────────────────────────────

> This is a presentation generator that uses AI
  to create reveal.js slides through an
  interactive Q&A process.

  It supports multiple themes and layouts.█

───────────────────────────────────

Enter for new line • Ctrl+D to submit • Esc to cancel
```

### 3. Options Selection (`OptionsModel`)

Multiple choice selection component.

```go
model := tui.NewOptionsInput(
    "What is your favorite language?",
    "Select one option from the list",
    []string{"Go", "Rust", "Python", "TypeScript", "Other"},
)

p := tea.NewProgram(model)
finalModel, _ := p.Run()
result := finalModel.(tui.OptionsModel)

selected := result.GetAnswer()          // string (option text)
index := result.GetSelectedIndex()      // int (0-based)
```

**Features:**

- Arrow keys (↑/↓) or vim keys (j/k) to navigate
- Number keys (1-9) for quick selection
- Enter or Space to confirm
- Radio button-style visual indicator

**Visual Example:**

```
What is your favorite language?
Select one option from the list

───────────────────────────────────

[ ] Go
[ ] Rust
[x] Python
[ ] TypeScript
[ ] Other

───────────────────────────────────

↑/↓ or j/k to select • Enter/Space to confirm • 1-9 for quick select • Esc to cancel
```

## Usage Patterns

### Sequential Inputs

```go
// Collect multiple inputs in sequence
inputs := []struct {
    prompt string
    iType  string
}{
    {"What is your name?", "text"},
    {"Do you want to continue?", "yesno"},
    {"Pick a theme:", "options"},
}

answers := make(map[string]string)

for _, input := range inputs {
    var model tea.Model

    switch input.iType {
    case "text":
        model = tui.NewTextInput(input.prompt, "", tui.SingleLine)
    case "yesno":
        model = tui.NewYesNoInput(input.prompt, "")
    case "options":
        model = tui.NewOptionsInput(input.prompt, "",
            []string{"Option 1", "Option 2", "Option 3"})
    }

    p := tea.NewProgram(model)
    finalModel, _ := p.Run()

    // Extract answer based on type
    // ... store in answers map
}
```

### Conditional Inputs

```go
// Ask yes/no, then follow up based on answer
continueModel := tui.NewYesNoInput(
    "Do you want advanced options?",
    "",
)
p := tea.NewProgram(continueModel)
finalModel, _ := p.Run()
result := finalModel.(tui.YesNoModel)

if result.GetAnswer() {
    // Ask for additional input
    detailsModel := tui.NewTextInput(
        "Describe your requirements:",
        "",
        tui.MultiLine,
    )
    p = tea.NewProgram(detailsModel)
    // ... handle response
}
```

### Integration with BAML

Perfect for gathering structured input for AI prompts:

```go
// 1. BAML generates input prompts
type InputPrompt struct {
    Prompt  string
    Type    string // "text", "yesno", "options"
    Options []string
}

prompts, _ := baml_client.GenerateInputPrompts(ctx, topic)

// 2. Present inputs with TUI components
responses := []string{}

for _, prompt := range prompts {
    var model tea.Model

    switch prompt.Type {
    case "text":
        model = tui.NewTextInput(prompt.Prompt, "", tui.SingleLine)
    case "yesno":
        model = tui.NewYesNoInput(prompt.Prompt, "")
    case "options":
        model = tui.NewOptionsInput(prompt.Prompt, "", prompt.Options)
    }

    p := tea.NewProgram(model)
    finalModel, _ := p.Run()

    // Collect response
    var answer string
    switch m := finalModel.(type) {
    case tui.TextModel:
        answer = m.GetAnswer()
    case tui.YesNoModel:
        answer = m.GetAnswerString()
    case tui.OptionsModel:
        answer = m.GetAnswer()
    }

    responses = append(responses, answer)
}

// 3. Send responses to AI for processing
result, _ := baml_client.ProcessResponses(ctx, responses)
```

## Styling

All components use the shared styles from `tui`:

```go
tui.TitleStyle      // Bold, prominent titles
tui.QuestionStyle   // Bold, colored questions
tui.HelpStyle       // Italic, subdued help text
tui.InputStyle      // Colored user input
tui.ErrorStyle      // Bold, red errors
tui.SuccessStyle    // Bold, green success
```

## Keyboard Shortcuts

All components support:

- `Ctrl+C` or `Esc` - Cancel/quit
- `Enter` - Confirm (context-dependent)

Component-specific:

- **Yes/No**: `y`/`n` for quick selection, `↑`/`↓` or `j`/`k` to navigate
- **Text (single)**: `Enter` to submit
- **Text (multi)**: `Enter` for newline, `Ctrl+D` to submit
- **Options**: `↑`/`↓` or `j`/`k` to navigate, `1-9` for quick select

## Advanced Features

### Text Wrapping

Both text input components automatically wrap text at word boundaries:

```go
// Text wraps at terminal width - 6
// Falls back to 40 chars minimum
// Wraps at spaces, not mid-word
// Continuation lines are indented
```

### Terminal Resize

Components automatically adjust to terminal size changes:

```go
// WindowSizeMsg is handled automatically
// Text wrapping updates dynamically
// No manual intervention needed
```

### State Management

Check if a question was answered:

```go
if model.IsDone() {
    answer := model.GetAnswer()
    // Process answer
}
```

## Examples

See `examples/input_components.go` for a complete demonstration:

```bash
go run examples/input_components.go
```

Or build and run:

```bash
make run-examples
# or manually
go build -o input_demo examples/input_components.go
./input_demo
```

## Testing

```bash
# Run with example program
make run-examples

# Test individual components
go test ./tui -v

# Build everything to verify compilation
go build ./...
```

## Migration from Custom Input

If you have custom input handling, here's how to migrate:

### Before (Custom Input):

```go
var input string
fmt.Print("Enter your name: ")
fmt.Scanln(&input)
```

### After (TUI Component):

```go
model := tui.NewTextInput(
    "What is your name?",
    "",
    tui.SingleLine,
)
p := tea.NewProgram(model)
finalModel, _ := p.Run()
result := finalModel.(tui.TextModel)
input := result.GetAnswer()
```

## Architecture

Each component:

1. Implements `tea.Model` interface
2. Handles its own state and rendering
3. Returns final state when done
4. Can be composed with other components
5. Uses consistent styling and UX patterns

## Future Enhancements

Planned features:

- **ValidationModel** - Input validation with custom rules
- **MultiSelectModel** - Select multiple options from list
- **DatePickerModel** - Date selection interface
- **RangeModel** - Numeric range selection
- **ConfirmModel** - Confirmation dialog with customizable buttons
- **FormModel** - Compose multiple inputs into a form

## License

MIT - Same as parent project
