package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/geoffjay/agar/tools"
)

// ToolsCommand lists all available tools in the tool registry
type ToolsCommand struct {
	toolRegistry *tools.ToolRegistry
}

// NewToolsCommand creates a new tools command with the given tool registry
func NewToolsCommand(toolRegistry *tools.ToolRegistry) *ToolsCommand {
	return &ToolsCommand{
		toolRegistry: toolRegistry,
	}
}

func (c *ToolsCommand) Name() string {
	return "tools"
}

func (c *ToolsCommand) Description() string {
	return "List all available tools in the tool registry"
}

func (c *ToolsCommand) Usage() string {
	return "/tools [tool-name]"
}

func (c *ToolsCommand) Aliases() []string {
	return []string{}
}

func (c *ToolsCommand) Execute(ctx context.Context, args []string, state ApplicationState) error {
	if c.toolRegistry == nil {
		state.AddLine("No tool registry available")
		return nil
	}

	// If a specific tool is requested, show detailed info
	if len(args) > 0 {
		toolName := args[0]
		tool, err := c.toolRegistry.Get(toolName)
		if err != nil {
			state.AddLine(fmt.Sprintf("Tool %q not found", toolName))
			return nil
		}

		state.AddLine(fmt.Sprintf("Tool: %s", tool.Name()))
		state.AddLine(fmt.Sprintf("Description: %s", tool.Description()))
		state.AddLine("")
		state.AddLine("Schema:")
		schema := tool.Schema()
		if schema != nil {
			// Pretty print the schema
			if properties, ok := schema["properties"].(map[string]interface{}); ok {
				for key, value := range properties {
					if propMap, ok := value.(map[string]interface{}); ok {
						desc := propMap["description"]
						required := false
						if req, ok := schema["required"].([]interface{}); ok {
							for _, r := range req {
								if r == key {
									required = true
									break
								}
							}
						}
						requiredStr := ""
						if required {
							requiredStr = " (required)"
						}
						state.AddLine(fmt.Sprintf("  %s%s: %v", key, requiredStr, desc))
					}
				}
			}
		}
		return nil
	}

	// Show all tools
	state.AddLine("Available Tools:")
	state.AddLine("")

	allTools := c.toolRegistry.ListTools()
	if len(allTools) == 0 {
		state.AddLine("No tools registered")
		return nil
	}

	// Sort tools by name
	sort.Slice(allTools, func(i, j int) bool {
		return allTools[i].Name() < allTools[j].Name()
	})

	// Find max name length for formatting
	maxLen := 0
	for _, tool := range allTools {
		if len(tool.Name()) > maxLen {
			maxLen = len(tool.Name())
		}
	}

	// Display each tool
	for _, tool := range allTools {
		padding := strings.Repeat(" ", maxLen-len(tool.Name())+2)
		line := fmt.Sprintf("  %s%s- %s", tool.Name(), padding, tool.Description())
		state.AddLine(line)
	}

	state.AddLine("")
	state.AddLine("Type '/tools <tool-name>' for more information about a specific tool")

	return nil
}
