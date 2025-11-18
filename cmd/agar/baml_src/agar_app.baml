// Agar Application Configuration Models

class AgarAppConfig {
  name string @description("Application name")
  description string @description("Application description")
  module_path string @description("Go module path (e.g., github.com/user/app)")
  tools ToolConfig[] @description("Tools to include in the application")
  components ComponentConfig[] @description("TUI components to include")
  features string[] @description("Additional features to implement")
  main_structure MainStructure @description("Main application structure")
}

class ToolConfig {
  name string @description("Tool name (read, write, search, etc.)")
  purpose string @description("Why this tool is needed")
  custom_config string? @description("Custom configuration if needed")
}

class ComponentConfig {
  type string @description("Component type (Application, Panel, Form, etc.)")
  name string @description("Component identifier")
  config string @description("Configuration details")
}

class MainStructure {
  use_tui bool @description("Whether to use TUI interface")
  entry_point string @description("Main entry point structure (tui, cli, hybrid)")
  initialization string[] @description("Initialization steps in order")
}

// Function to create Agar application configuration
function CreateAgarApp(requirements: string, project_name: string) -> AgarAppConfig {
  client CustomSonnet4
  prompt #"
    You are an expert in building AI agent applications using the Agar framework.

    Agar provides:
    - Tools framework with 11 tools: read, write, delete, list, glob, fetch, download, search, grep, shell, tasklist
    - TUI components: Application, Panel, Footer, Layout, Input components, Iterative forms

    Based on these requirements:
    {{ requirements }}

    Create a complete configuration for a new Agar application named "{{ project_name }}".

    Guidelines:
    1. Select appropriate tools based on what the application needs to do
    2. Choose TUI components if it needs user interaction
    3. Suggest features that would be valuable
    4. Provide clear initialization steps
    5. Use a module path like: github.com/user/{{ project_name }}

    {{ ctx.output_format }}
  "#
}

// Function to generate project structure details
function GenerateProjectStructure(config: AgarAppConfig) -> ProjectStructure {
  client CustomSonnet4
  prompt #"
    Given this Agar application configuration:

    Name: {{ config.name }}
    Description: {{ config.description }}
    Tools: {{ config.tools }}
    Components: {{ config.components }}
    Features: {{ config.features }}

    Generate a detailed project structure including:
    1. Directory layout
    2. File names and their purposes
    3. Package organization
    4. Dependencies needed

    {{ ctx.output_format }}
  "#
}

class ProjectStructure {
  directories string[] @description("Directories to create")
  files FileSpec[] @description("Files to generate")
  dependencies string[] @description("Go dependencies to add")
}

class FileSpec {
  path string @description("File path relative to project root")
  content string @description("File content")
  purpose string @description("What this file does")
}

// Function to generate main.go content
function GenerateMainGo(config: AgarAppConfig) -> string {
  client CustomSonnet4
  prompt #"
    Generate a main.go file for this Agar application:

    Name: {{ config.name }}
    Description: {{ config.description }}
    Use TUI: {{ config.main_structure.use_tui }}
    Entry Point: {{ config.main_structure.entry_point }}
    Tools: {{ config.tools }}
    Components: {{ config.components }}

    The main.go should:
    1. Import necessary agar packages (tools, tui)
    2. Initialize a tool registry with the specified tools
    3. Create TUI application if use_tui is true
    4. Set up proper error handling
    5. Follow Go best practices

    Return ONLY the Go source code, no explanations.
  "#
}

// Response from the AI agent
class AgentResponse {
  response string @description("The agent's response to the user")
  suggestions string[] @description("Optional follow-up suggestions or actions")
}

// Function to handle agent prompts in the TUI
function AgentPrompt(user_input: string, conversation_history: ConversationMessage[]) -> AgentResponse {
  client CustomSonnet4
  prompt #"
    You are an AI assistant integrated into the Agar CLI application.

    Agar is a Go framework for building AI agent applications with:
    - **Tools Framework**: 11 built-in tools (read, write, delete, list, glob, fetch, download, search, grep, shell, tasklist)
    - **TUI Components**: Terminal UI with Bubble Tea (Application, Panel, Footer, Layout, Input components)
    - **Command System**: Slash commands for interactive operations
    - **Tool Registry**: Thread-safe tool management with dynamic registration

    Your role is to:
    1. Help users understand and use Agar framework features
    2. Assist with Go programming and best practices
    3. Provide guidance on building AI agent applications
    4. Answer questions about the available tools and TUI components
    5. Help debug issues and suggest solutions

    Available slash commands users can type:
    - /help - Show available commands
    - /tools - List all available tools
    - /init <name> - Create a new Agar project
    - /exit - Exit the application
    - /clear - Clear the screen
    - /export <file> - Export conversation
    - /import <file> - Import conversation

    {% if conversation_history|length > 0 %}
    Conversation history:
    {% for msg in conversation_history %}
    {{ msg.role }}: {{ msg.content }}
    {% endfor %}
    {% endif %}

    User: {{ user_input }}

    Provide a helpful, concise response. If you suggest code, use Go syntax.
    If the user's question could be solved with a slash command, mention it.
    Be friendly and encouraging.

    {{ ctx.output_format }}
  "#
}

// Conversation message for maintaining context
class ConversationMessage {
  role string @description("Role: 'user' or 'assistant'")
  content string @description("Message content")
}
