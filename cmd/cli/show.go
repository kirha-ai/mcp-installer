package cli

import (
	"github.com/kirha-ai/mcp-installer/internal/core/domain/installer"
	"github.com/spf13/cobra"
)

func NewCmdShow() *cobra.Command {
	var (
		client     string
		vertical   string
		configPath string
		onlyKirha  bool
		verbose    bool
	)

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display current MCP server configuration for a client",
		Long: `Display the current MCP server configuration for the specified development environment.

This command will show the existing MCP server configuration, including any Kirha MCP servers
and other MCP servers that are configured. API keys will be masked for security.`,
		Example: `  # Show all MCP server configurations for Claude Desktop
  mcp-installer show --client claude

  # Show only crypto vertical configuration for Docker
  mcp-installer show --client docker --vertical crypto

  # Show only Kirha MCP servers for Claude Desktop
  mcp-installer show --client claude --only-kirha

  # Show all configurations for VS Code with verbose output
  mcp-installer show --client vscode --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(cmd, installer.OperationShow, client, vertical, "", configPath, false, verbose, onlyKirha)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client to show configuration for (required)")
	cmd.Flags().StringVar(&vertical, "vertical", "", "Vertical to show (crypto, utils) (optional - shows all if not specified)")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path (optional)")
	cmd.Flags().BoolVar(&onlyKirha, "only-kirha", false, "Show only Kirha MCP servers")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	_ = cmd.MarkFlagRequired("client")

	return cmd
}
