package cli

import (
	"github.com/spf13/cobra"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
)

func NewCmdInstall() *cobra.Command {
	var (
		client     string
		apiKey     string
		configPath string
		dryRun     bool
		verbose    bool
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Kirha MCP server for a client",
		Long: `Install Kirha MCP server for the specified development environment.

This command will add the Kirha MCP server to the client's configuration.
If the server already exists, the command will fail with a suggestion to use 'update' instead.`,
		Example: `  # Install for Claude Code CLI
  mcp-installer install --client claudecode --key your-api-key-here

  # Install for Cursor with dry run
  mcp-installer install --client cursor --key your-api-key-here --dry-run

  # Install for Codex with custom config path
  mcp-installer install --client codex --key your-api-key-here --config-path /custom/path

  # Install for OpenCode with verbose output
  mcp-installer install --client opencode --key your-api-key-here --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(cmd, installer.OperationInstall, client, apiKey, configPath, dryRun, verbose, force)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client to install for (claudecode, cursor, codex, opencode) (required)")
	cmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key for Kirha MCP server (required)")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path (optional)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force installation even if the client is running")

	_ = cmd.MarkFlagRequired("client")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}
