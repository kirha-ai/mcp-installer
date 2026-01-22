package cli

import (
	"github.com/spf13/cobra"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
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
		Example: `  # Show MCP server configuration for Claude Code CLI
  mcp-installer show --client claudecode

  # Show configuration for Codex with verbose output
  mcp-installer show --client codex --verbose

  # Show configuration for OpenCode
  mcp-installer show --client opencode

  # Show configuration for Droid (Factory AI)
  mcp-installer show --client droid`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(cmd, installer.OperationShow, client, "", configPath, false, verbose, false)
		},
	}

	cmd.Flags().StringVarP(&client, "client", "c", "", "Client to show configuration for (claudecode, codex, opencode, gemini, droid) (required)")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path (optional)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	_ = cmd.MarkFlagRequired("client")

	return cmd
}
