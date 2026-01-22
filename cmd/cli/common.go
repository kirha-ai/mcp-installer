package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.kirha.ai/mcp-installer/di"
	domainErrors "go.kirha.ai/mcp-installer/internal/core/domain/errors"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
)

func runOperation(cmd *cobra.Command, operation installer.OperationType, client, apiKey, configPath string, dryRun, verbose, force bool) error {
	clientType, err := validateClient(client)
	if err != nil {
		return err
	}

	if operation == installer.OperationInstall && apiKey == "" {
		return fmt.Errorf("API key is required for %s operation", operation)
	}

	config := &installer.Config{
		Client:     clientType,
		ApiKey:     apiKey,
		ConfigPath: configPath,
		Operation:  operation,
		DryRun:     dryRun,
		Verbose:    verbose,
		Force:      force,
	}

	app, err := di.ProvideInstallerApplication()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	result, err := app.Execute(ctx, config)

	if err != nil {
		if errors.Is(err, domainErrors.ErrServerExistsUseUpdate) {
			return fmt.Errorf("MCP server already exists for %s. Use 'mcp-installer update --client %s --key <api-key>' to update it", client, client)
		} else if errors.Is(err, domainErrors.ErrServerNotFoundForUpdate) {
			return fmt.Errorf("MCP server not found for %s. Use 'mcp-installer install --client %s --key <api-key>' to install it first", client, client)
		} else if errors.Is(err, domainErrors.ErrServerNotFoundForRemove) {
			return fmt.Errorf("MCP server not found for %s. Nothing to remove", client)
		} else if errors.Is(err, domainErrors.ErrClientRunning) {
			return fmt.Errorf("the %s application is currently running. Please close it and try again", client)
		} else if errors.Is(err, domainErrors.ErrUnsupportedClient) {
			return fmt.Errorf("unsupported client: %s\n\nSupported clients: claudecode, codex, opencode, gemini, droid", client)
		} else {
			return fmt.Errorf("operation failed: %w", err)
		}
	}

	fmt.Println(result.Message)

	if verbose && result.ConfigPath != "" {
		fmt.Printf("\nConfiguration file: %s\n", result.ConfigPath)
		if result.BackupPath != "" {
			fmt.Printf("Backup created at: %s\n", result.BackupPath)
		}
	}

	return nil
}

func validateClient(client string) (installer.ClientType, error) {
	switch strings.ToLower(client) {
	case "claudecode", "claude-code":
		return installer.ClientTypeClaudecode, nil
	case "codex":
		return installer.ClientTypeCodex, nil
	case "opencode":
		return installer.ClientTypeOpencode, nil
	case "gemini", "gemini-cli":
		return installer.ClientTypeGemini, nil
	case "droid", "factory":
		return installer.ClientTypeDroid, nil
	default:
		return "", domainErrors.ErrUnsupportedClient
	}
}
