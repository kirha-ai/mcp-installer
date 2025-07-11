package cli

import (
	"fmt"
	
	"github.com/spf13/cobra"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
)

func NewCmdUpdate() *cobra.Command {
	var (
		client          string
		vertical        string
		apiKey          string
		configPath      string
		dryRun          bool
		verbose         bool
		enablePlanMode  bool
		disablePlanMode bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Kirha MCP server configuration for a client",
		Long: `Update the Kirha MCP server configuration for the specified development environment.

This command will update the existing MCP server configuration with new settings.
If the server doesn't exist, the command will fail with a suggestion to use 'install' instead.`,
		Example: `  # Update crypto vertical configuration for Claude Desktop
  mcp-installer update --client claude --vertical crypto --key your-new-api-key

  # Update utils vertical for Cursor with dry run
  mcp-installer update --client cursor --vertical utils --key your-new-api-key --dry-run

  # Update crypto vertical for Docker with verbose output
  mcp-installer update --client docker --vertical crypto --key your-new-api-key --verbose

  # Update crypto vertical with plan mode enabled
  mcp-installer update --client vscode --vertical crypto --key your-new-api-key --enable-plan-mode

  # Enable plan mode without changing API key
  mcp-installer update --client claude --vertical crypto --enable-plan-mode

  # Disable plan mode without changing API key
  mcp-installer update --client claude --vertical crypto --disable-plan-mode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if enablePlanMode && disablePlanMode {
				return fmt.Errorf("cannot use both --enable-plan-mode and --disable-plan-mode at the same time")
			}
			
			var planModeEnabled bool
			if enablePlanMode {
				planModeEnabled = true
			} else if disablePlanMode {
				planModeEnabled = false
			}
			
			return runOperation(cmd, installer.OperationUpdate, client, vertical, apiKey, configPath, dryRun, verbose, false, enablePlanMode, disablePlanMode, planModeEnabled)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client to update configuration for (required)")
	cmd.Flags().StringVar(&vertical, "vertical", "", "Vertical to update (crypto, utils) (required)")
	cmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key for Kirha MCP server (optional for plan mode changes)")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path (optional)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	cmd.Flags().BoolVar(&enablePlanMode, "enable-plan-mode", false, "Enable tool plan mode")
	cmd.Flags().BoolVar(&disablePlanMode, "disable-plan-mode", false, "Disable tool plan mode")

	_ = cmd.MarkFlagRequired("client")
	_ = cmd.MarkFlagRequired("vertical")

	return cmd
}
