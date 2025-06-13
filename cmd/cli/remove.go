package cli

import (
	"github.com/kirha-ai/mcp-installer/internal/core/domain/installer"
	"github.com/spf13/cobra"
)

func NewCmdRemove() *cobra.Command {
	var (
		client     string
		vertical   string
		configPath string
		dryRun     bool
		verbose    bool
	)

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove Kirha MCP server from a client",
		Long: `Remove the Kirha MCP server configuration from the specified development environment.

This command will completely remove the MCP server from the client's configuration.
If the server doesn't exist, the command will fail with an appropriate message.`,
		Example: `  # Remove crypto vertical from Claude Desktop
  mcp-installer remove --client claude --vertical crypto

  # Remove utils vertical from Docker with dry run
  mcp-installer remove --client docker --vertical utils --dry-run

  # Remove crypto vertical from VS Code with verbose output
  mcp-installer remove --client vscode --vertical crypto --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(cmd, installer.OperationRemove, client, vertical, "", configPath, dryRun, verbose, false)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client to remove MCP server from (required)")
	cmd.Flags().StringVar(&vertical, "vertical", "", "Vertical to remove (crypto, utils) (required)")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path (optional)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	_ = cmd.MarkFlagRequired("client")
	_ = cmd.MarkFlagRequired("vertical")

	return cmd
}
