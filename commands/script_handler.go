package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ScriptHandler executes command scripts with proper environment setup
type ScriptHandler struct {
	script string
	env    map[string]string
}

// Execute runs the script with the given application state
func (s *ScriptHandler) Execute(ctx context.Context, state ApplicationState) error {
	if s.script == "" {
		return fmt.Errorf("no script provided")
	}

	// Create temporary script file
	tmpFile, err := os.CreateTemp("", "agar-command-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp script: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write script to file
	if _, err := tmpFile.WriteString(s.script); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp script: %w", err)
	}

	// Make script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Build environment
	envVars := os.Environ()
	for k, v := range s.env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	// Execute script
	cmd := exec.CommandContext(ctx, "/bin/sh", tmpFile.Name())
	cmd.Env = envVars

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script execution failed: %w\nOutput: %s", err, string(output))
	}

	// Add output to application state if any
	if len(output) > 0 {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			state.AddLine(line)
		}
	}

	return nil
}
