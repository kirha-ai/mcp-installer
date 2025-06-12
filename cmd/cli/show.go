package cli

import (
	"github.com/kirha-ai/mcp-installer/internal/core/domain/installer"
	"github.com/spf13/cobra"
)

func NewCmdShow() *cobra.Command {
	var (
		client     string
		configPath string
		verbose    bool
	)

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display current MCP server configuration for a client",
		Long: `Display the current MCP server configuration for the specified development environment.

This command will show the existing MCP server configuration, including any Kirha MCP servers
and other MCP servers that are configured. API keys will be masked for security.`,
		Example: `  # Show configuration for Claude Desktop
  mcp-installer show --client claude

  # Show configuration for Docker
  mcp-installer show --client docker

  # Show configuration for VS Code with verbose output
  mcp-installer show --client vscode --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(cmd, installer.OperationShow, client, "", configPath, false, verbose)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client to show configuration for (required)")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path (optional)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	cmd.MarkFlagRequired("client")

	return cmd
}
