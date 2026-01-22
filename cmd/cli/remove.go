package cli

import (
	"github.com/spf13/cobra"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
)

func NewCmdRemove() *cobra.Command {
	var (
		client     string
		configPath string
		dryRun     bool
		verbose    bool
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove Kirha MCP server from a client",
		Long: `Remove the Kirha MCP server configuration from the specified development environment.

This command will completely remove the MCP server from the client's configuration.
If the server doesn't exist, the command will fail with an appropriate message.`,
		Example: `  # Remove from Claude Code CLI
  mcp-installer remove --client claudecode

  # Remove from Cursor with dry run
  mcp-installer remove --client cursor --dry-run

  # Remove from Codex with verbose output
  mcp-installer remove --client codex --verbose

  # Remove from OpenCode
  mcp-installer remove --client opencode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(cmd, installer.OperationRemove, client, "", configPath, dryRun, verbose, force)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client to remove MCP server from (claudecode, codex, opencode, gemini) (required)")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path (optional)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force removal even if the client is running")

	_ = cmd.MarkFlagRequired("client")

	return cmd
}
