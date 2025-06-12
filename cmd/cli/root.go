package cli

import (
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp-installer",
		Short: "MCP Installer - Install Kirha MCP server for various development environments",
		Long: `MCP Installer is a CLI tool that simplifies the installation of Kirha MCP 
(Model Context Protocol) server across multiple development environments.

Supported clients:
  - claude      Claude Desktop application
  - cursor      Cursor IDE
  - vscode      Visual Studio Code
  - claude-code Claude Code CLI tool
  - docker      Docker Compose setup`,
		Example: `  # Install for Claude Desktop
  mcp-installer install --client claude --key your-api-key-here

  # Update configuration for Docker
  mcp-installer update --client docker --key your-new-api-key

  # Remove from Cursor
  mcp-installer remove --client cursor

  # Show current configuration
  mcp-installer show --client claude

  # Dry run to see what would be changed
  mcp-installer install --client vscode --key your-api-key-here --dry-run`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(NewCmdInstall())
	cmd.AddCommand(NewCmdUpdate())
	cmd.AddCommand(NewCmdRemove())
	cmd.AddCommand(NewCmdShow())
	cmd.AddCommand(NewCmdVersion())

	return cmd
}
