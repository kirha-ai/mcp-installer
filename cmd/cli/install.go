package cli

import (
	"github.com/spf13/cobra"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
)

func NewCmdInstall() *cobra.Command {
	var (
		client         string
		vertical       string
		apiKey         string
		configPath     string
		dryRun         bool
		verbose        bool
		enablePlanMode bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Kirha MCP server for a client",
		Long: `Install Kirha MCP server for the specified development environment.

This command will add the Kirha MCP server to the client's configuration.
If the server already exists, the command will fail with a suggestion to use 'update' instead.`,
		Example: `  # Install crypto vertical for Claude Desktop
  mcp-installer install --client claude --vertical crypto --key your-api-key-here

  # Install utils vertical for Docker with dry run
  mcp-installer install --client docker --vertical utils --key your-api-key-here --dry-run

  # Install crypto vertical for VS Code with custom config path
  mcp-installer install --client vscode --vertical crypto --key your-api-key-here --config-path /custom/path

  # Install crypto vertical with plan mode enabled
  mcp-installer install --client cursor --vertical crypto --key your-api-key-here --enable-plan-mode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(cmd, installer.OperationInstall, client, vertical, apiKey, configPath, dryRun, verbose, false, enablePlanMode, false, enablePlanMode)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client to install for (required)")
	cmd.Flags().StringVar(&vertical, "vertical", "", "Vertical to install (crypto, utils) (required)")
	cmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key for Kirha MCP server (required)")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path (optional)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	cmd.Flags().BoolVar(&enablePlanMode, "enable-plan-mode", false, "Enable tool plan mode (default: false)")

	_ = cmd.MarkFlagRequired("client")
	_ = cmd.MarkFlagRequired("vertical")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}
