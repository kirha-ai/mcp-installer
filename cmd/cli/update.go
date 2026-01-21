package cli

import (
	"github.com/spf13/cobra"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
)

func NewCmdUpdate() *cobra.Command {
	var (
		client     string
		apiKey     string
		configPath string
		dryRun     bool
		verbose    bool
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Kirha MCP server configuration for a client",
		Long: `Update the Kirha MCP server configuration for the specified development environment.

This command will update the existing MCP server configuration with new settings.
If the server doesn't exist, the command will fail with a suggestion to use 'install' instead.`,
		Example: `  # Update configuration for Claude Code CLI
  mcp-installer update --client claudecode --key your-new-api-key

  # Update for Cursor with dry run
  mcp-installer update --client cursor --key your-new-api-key --dry-run

  # Update for Codex with verbose output
  mcp-installer update --client codex --key your-new-api-key --verbose

  # Update for OpenCode
  mcp-installer update --client opencode --key your-new-api-key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(cmd, installer.OperationUpdate, client, apiKey, configPath, dryRun, verbose, force)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client to update configuration for (claudecode, cursor, codex, opencode) (required)")
	cmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key for Kirha MCP server (optional - preserves existing if not provided)")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path (optional)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force update even if the client is running")

	_ = cmd.MarkFlagRequired("client")

	return cmd
}
