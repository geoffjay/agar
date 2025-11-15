package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ShellTool implements safe shell command execution
type ShellTool struct{}

// ShellParams defines the parameters for the Shell tool
type ShellParams struct {
	Command     string            `json:"command"`
	Args        []string          `json:"args,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Timeout     int               `json:"timeout,omitempty"` // seconds
	Shell       string            `json:"shell,omitempty"`   // "bash", "sh", "powershell"
}

// ShellResult represents the result of a shell command execution
type ShellResult struct {
	Command  string `json:"command"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Duration int64  `json:"duration_ms"` // Duration in milliseconds
	Timeout  bool   `json:"timeout,omitempty"`
}

// NewShellTool creates a new Shell tool instance
func NewShellTool() *ShellTool {
	return &ShellTool{}
}

// Name returns the tool's name
func (t *ShellTool) Name() string {
	return "shell"
}

// Description returns the tool's description
func (t *ShellTool) Description() string {
	return "Execute shell commands with timeout support, environment variables, and working directory specification"
}

// Schema returns the JSON schema for the tool's parameters
func (t *ShellTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Command to execute",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"description": "Command arguments",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for command execution",
			},
			"environment": map[string]interface{}{
				"type":        "object",
				"description": "Environment variables to set",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 30, max: 300)",
				"minimum":     1,
				"maximum":     300,
			},
			"shell": map[string]interface{}{
				"type":        "string",
				"description": "Shell to use: 'bash', 'sh', or 'powershell'",
				"enum":        []string{"bash", "sh", "powershell"},
			},
		},
		"required": []string{"command"},
	}
}

// Validate checks if the parameters are valid
func (t *ShellTool) Validate(params json.RawMessage) error {
	var p ShellParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Command == "" {
		return fmt.Errorf("command is required")
	}

	// Security: Block potentially dangerous commands
	dangerous := []string{"rm -rf /", ":(){ :|:& };:", "mkfs", "dd if=/dev/zero"}
	for _, d := range dangerous {
		if strings.Contains(p.Command, d) {
			return fmt.Errorf("command contains potentially dangerous pattern: %s", d)
		}
	}

	if p.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative")
	}

	if p.Timeout > 300 {
		return fmt.Errorf("timeout must not exceed 300 seconds")
	}

	if p.Shell != "" && p.Shell != "bash" && p.Shell != "sh" && p.Shell != "powershell" {
		return fmt.Errorf("shell must be 'bash', 'sh', or 'powershell'")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *ShellTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p ShellParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Default timeout to 30 seconds
	timeout := 30
	if p.Timeout > 0 {
		timeout = p.Timeout
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Prepare command
	var cmd *exec.Cmd
	if p.Shell != "" {
		// Execute through specified shell
		shellCmd := t.getShellCommand(p.Shell, p.Command)
		cmd = exec.CommandContext(execCtx, shellCmd[0], shellCmd[1:]...)
	} else {
		// Execute command directly
		if len(p.Args) > 0 {
			cmd = exec.CommandContext(execCtx, p.Command, p.Args...)
		} else {
			// If no args, try to parse command string
			parts := strings.Fields(p.Command)
			if len(parts) > 1 {
				cmd = exec.CommandContext(execCtx, parts[0], parts[1:]...)
			} else {
				cmd = exec.CommandContext(execCtx, p.Command)
			}
		}
	}

	// Set working directory
	if p.WorkingDir != "" {
		cmd.Dir = p.WorkingDir
	}

	// Set environment variables
	if len(p.Environment) > 0 {
		env := cmd.Environ()
		for key, value := range p.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command and measure duration
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	result := &ShellResult{
		Command:  p.Command,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration.Milliseconds(),
	}

	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded {
		result.Timeout = true
		result.ExitCode = -1
		return result, nil
	}

	// Get exit code
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	} else {
		result.ExitCode = 0
	}

	return result, nil
}

// getShellCommand returns the shell command array for the given shell type
func (t *ShellTool) getShellCommand(shell, command string) []string {
	switch shell {
	case "bash":
		return []string{"bash", "-c", command}
	case "sh":
		return []string{"sh", "-c", command}
	case "powershell":
		return []string{"powershell", "-Command", command}
	default:
		return []string{"sh", "-c", command}
	}
}
