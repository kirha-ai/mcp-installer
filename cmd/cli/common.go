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

func runOperation(cmd *cobra.Command, operation installer.OperationType, client, vertical, apiKey, configPath string, dryRun, verbose, onlyKirha bool, enablePlanMode, disablePlanMode bool, planModeValue bool) error {
	clientType, err := validateClient(client)
	if err != nil {
		return err
	}

	var verticalType installer.VerticalType
	if vertical != "" {
		verticalType, err = validateVertical(vertical)
		if err != nil {
			return err
		}
	}

	if operation == installer.OperationInstall && apiKey == "" {
		return fmt.Errorf("API key is required for %s operation", operation)
	}

	config := &installer.Config{
		Client:          clientType,
		Vertical:        verticalType,
		ApiKey:          apiKey,
		ConfigPath:      configPath,
		Operation:       operation,
		DryRun:          dryRun,
		Verbose:         verbose,
		OnlyKirha:       onlyKirha,
		EnablePlanMode:  enablePlanMode,
		DisablePlanMode: disablePlanMode,
		PlanModeSet:     enablePlanMode || disablePlanMode,
	}

	app, err := di.ProvideInstallerApplication()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	result, err := app.Execute(ctx, config)

	if err != nil {
		if errors.Is(err, domainErrors.ErrServerExistsUseUpdate) {
			return fmt.Errorf("MCP server already exists for %s %s vertical. Use 'mcp-installer update --client %s --vertical %s --key <api-key>' to update it", client, vertical, client, vertical)
		} else if errors.Is(err, domainErrors.ErrServerNotFoundForUpdate) {
			return fmt.Errorf("MCP server not found for %s %s vertical. Use 'mcp-installer install --client %s --vertical %s --key <api-key>' to install it first", client, vertical, client, vertical)
		} else if errors.Is(err, domainErrors.ErrServerNotFoundForRemove) {
			return fmt.Errorf("MCP server not found for %s %s vertical. Nothing to remove", client, vertical)
		} else if errors.Is(err, domainErrors.ErrClientRunning) {
			return fmt.Errorf("the %s application is currently running. Please close it and try again", client)
		} else if errors.Is(err, domainErrors.ErrUnsupportedClient) {
			return fmt.Errorf("unsupported client: %s\n\nSupported clients: claude, cursor, vscode, claude-code, docker", client)
		} else if errors.Is(err, domainErrors.ErrUnsupportedVertical) {
			return fmt.Errorf("unsupported vertical: %s\n\nSupported verticals: crypto, utils", vertical)
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
	case "claude":
		return installer.ClientTypeClaude, nil
	case "cursor":
		return installer.ClientTypeCursor, nil
	case "vscode", "vs-code", "code":
		return installer.ClientTypeVSCode, nil
	case "claude-code", "claudecode":
		return installer.ClientTypeClaudeCode, nil
	case "docker":
		return installer.ClientTypeDocker, nil
	default:
		return "", domainErrors.ErrUnsupportedClient
	}
}

func validateVertical(vertical string) (installer.VerticalType, error) {
	switch strings.ToLower(vertical) {
	case "crypto":
		return installer.VerticalTypeCrypto, nil
	default:
		return "", domainErrors.ErrUnsupportedVertical
	}
}
