package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func NewCmdRoot() *cobra.Command {
	var versionFlag bool

	cmd := &cobra.Command{
		Use:   "mcp-installer",
		Short: "MCP Installer - Install Kirha MCP server for various development environments",
		Long: `MCP Installer is a CLI tool that simplifies the installation of Kirha MCP
(Model Context Protocol) server across multiple development environments.

Supported clients:
  - claudecode  Claude Code CLI tool
  - cursor      Cursor IDE
  - codex       OpenAI Codex CLI
  - opencode    OpenCode IDE`,
		Example: `  # Install for Claude Code CLI
  mcp-installer install --client claudecode --key your-api-key-here

  # Update configuration for Cursor
  mcp-installer update --client cursor --key your-new-api-key

  # Remove from Codex
  mcp-installer remove --client codex

  # Show current configuration
  mcp-installer show --client claudecode

  # Dry run to see what would be changed
  mcp-installer install --client opencode --key your-api-key-here --dry-run`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if versionFlag {
				fmt.Printf("Kirha MCP Gateway version %s\n", Version)
				os.Exit(0)
			}

			// Check for updates on every command execution
			checkForUpdates()
		},
	}
	cmd.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "display version information")

	cmd.AddCommand(NewCmdInstall())
	cmd.AddCommand(NewCmdUpdate())
	cmd.AddCommand(NewCmdRemove())
	cmd.AddCommand(NewCmdShow())
	cmd.AddCommand(NewCmdVersion())
	cmd.AddCommand(NewCmdUpdateVersion())

	return cmd
}
