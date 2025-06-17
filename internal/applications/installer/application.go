package installer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go.kirha.ai/mcp-installer/internal/core/domain/errors"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
	"go.kirha.ai/mcp-installer/internal/core/ports"
	"go.kirha.ai/mcp-installer/internal/core/ports/factories"
)

type Application struct {
	installerFactory factories.InstallerFactory
}

func New(installerFactory factories.InstallerFactory) *Application {
	return &Application{
		installerFactory: installerFactory,
	}
}

func (a *Application) Execute(ctx context.Context, config *installer.Config) (*installer.InstallResult, error) {
	switch config.Operation {
	case installer.OperationInstall:
		return a.install(ctx, config)
	case installer.OperationUpdate:
		return a.update(ctx, config)
	case installer.OperationRemove:
		return a.remove(ctx, config)
	case installer.OperationShow:
		showResult, err := a.show(ctx, config)
		if err != nil {
			return nil, err
		}
		return &installer.InstallResult{
			Success:    showResult.Success,
			ConfigPath: showResult.ConfigPath,
			Message:    showResult.Message,
		}, nil
	default:
		slog.ErrorContext(ctx, "unknown operation requested",
			slog.String("operation", string(config.Operation)))
		return nil, errors.ErrUnknownOperation
	}
}

func (a *Application) install(ctx context.Context, config *installer.Config) (*installer.InstallResult, error) {
	slog.InfoContext(ctx, "starting installation",
		slog.String("client", string(config.Client)),
		slog.Bool("dry_run", config.DryRun))

	if err := a.validateApiKey(config.ApiKey); err != nil {
		return nil, err
	}

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get installer for client",
			slog.String("error", err.Error()),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	running, err := clientInstaller.IsClientRunning(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to check if client is running", slog.String("error", err.Error()))
	}
	if running && !config.DryRun {
		slog.WarnContext(ctx, "client is currently running",
			slog.String("client", string(config.Client)))
		return nil, errors.ErrClientRunning
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get config path", slog.String("error", err.Error()))
		return nil, err
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load config", slog.String("error", err.Error()))
		return nil, err
	}

	exists, err := clientInstaller.HasMcpServer(ctx, currentConfig, config.Vertical)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check if server exists", slog.String("error", err.Error()))
		return nil, err
	}

	if exists {
		slog.ErrorContext(ctx, "MCP server already exists, use 'update' command to modify it")
		return nil, errors.ErrServerExistsUseUpdate
	}

	if config.DryRun {
		slog.InfoContext(ctx, "dry run - would install server",
			slog.String("path", configPath))
		return &installer.InstallResult{
			Success:    true,
			ConfigPath: configPath,
			Message:    fmt.Sprintf("Would install Kirha MCP server to %s", configPath),
		}, nil
	}

	return a.performInstallOrUpdate(ctx, config, currentConfig, clientInstaller, "installed")
}

func (a *Application) update(ctx context.Context, config *installer.Config) (*installer.InstallResult, error) {
	slog.InfoContext(ctx, "starting update",
		slog.String("client", string(config.Client)),
		slog.Bool("dry_run", config.DryRun))

	if err := a.validateApiKey(config.ApiKey); err != nil {
		return nil, err
	}

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get installer for client",
			slog.String("error", err.Error()),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	running, err := clientInstaller.IsClientRunning(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to check if client is running", slog.String("error", err.Error()))
	}
	if running && !config.DryRun {
		slog.WarnContext(ctx, "client is currently running",
			slog.String("client", string(config.Client)))
		return nil, errors.ErrClientRunning
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get config path", slog.String("error", err.Error()))
		return nil, err
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load config", slog.String("error", err.Error()))
		return nil, err
	}

	exists, err := clientInstaller.HasMcpServer(ctx, currentConfig, config.Vertical)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check if server exists", slog.String("error", err.Error()))
		return nil, err
	}

	if !exists {
		slog.ErrorContext(ctx, "MCP server not found, use 'install' command to add it")
		return nil, errors.ErrServerNotFoundForUpdate
	}

	if config.DryRun {
		slog.InfoContext(ctx, "dry run - would update server",
			slog.String("path", configPath))
		return &installer.InstallResult{
			Success:    true,
			ConfigPath: configPath,
			Message:    fmt.Sprintf("Would update Kirha MCP server in %s", configPath),
		}, nil
	}

	configWithoutServer, err := clientInstaller.RemoveMcpServer(ctx, currentConfig, config.Vertical)
	if err != nil {
		slog.ErrorContext(ctx, "failed to remove existing server", slog.String("error", err.Error()))
		return nil, err
	}

	return a.performInstallOrUpdate(ctx, config, configWithoutServer, clientInstaller, "updated")
}

func (a *Application) remove(ctx context.Context, config *installer.Config) (*installer.InstallResult, error) {
	slog.InfoContext(ctx, "starting removal",
		slog.String("client", string(config.Client)),
		slog.Bool("dry_run", config.DryRun))

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get installer for client",
			slog.String("error", err.Error()),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	running, err := clientInstaller.IsClientRunning(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to check if client is running", slog.String("error", err.Error()))
	}
	if running && !config.DryRun {
		slog.WarnContext(ctx, "client is currently running",
			slog.String("client", string(config.Client)))
		return nil, errors.ErrClientRunning
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get config path", slog.String("error", err.Error()))
		return nil, err
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load config", slog.String("error", err.Error()))
		return nil, err
	}

	exists, err := clientInstaller.HasMcpServer(ctx, currentConfig, config.Vertical)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check if server exists", slog.String("error", err.Error()))
		return nil, err
	}

	if !exists {
		slog.ErrorContext(ctx, "MCP server not found, nothing to remove")
		return nil, errors.ErrServerNotFoundForRemove
	}

	if config.DryRun {
		slog.InfoContext(ctx, "dry run - would remove server",
			slog.String("path", configPath))
		return &installer.InstallResult{
			Success:    true,
			ConfigPath: configPath,
			Message:    fmt.Sprintf("Would remove Kirha MCP server from %s", configPath),
		}, nil
	}

	backupPath, err := clientInstaller.BackupConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create backup", slog.String("error", err.Error()))
	}

	updatedConfig, err := clientInstaller.RemoveMcpServer(ctx, currentConfig, config.Vertical)
	if err != nil {
		slog.ErrorContext(ctx, "failed to remove MCP server", slog.String("error", err.Error()))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				slog.ErrorContext(ctx, "failed to restore backup", slog.String("error", restoreErr.Error()))
			}
		}

		return nil, err
	}

	if err := clientInstaller.SaveConfig(ctx, updatedConfig); err != nil {
		slog.ErrorContext(ctx, "failed to save config", slog.String("error", err.Error()))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				slog.ErrorContext(ctx, "failed to restore backup", slog.String("error", restoreErr.Error()))
			}
		}

		return nil, err
	}

	message := fmt.Sprintf("Successfully removed Kirha MCP server from %s", config.Client)
	if running {
		message += ". Please restart the application to apply changes."
	}

	slog.InfoContext(ctx, "removal completed successfully",
		slog.String("config_path", configPath),
		slog.String("backup_path", backupPath))

	return &installer.InstallResult{
		Success:    true,
		ConfigPath: configPath,
		BackupPath: backupPath,
		Message:    message,
	}, nil
}

func (a *Application) performInstallOrUpdate(ctx context.Context, config *installer.Config, currentConfig interface{}, clientInstaller ports.Installer, operation string) (*installer.InstallResult, error) {
	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get config path", slog.String("error", err.Error()))
		return nil, err
	}

	backupPath, err := clientInstaller.BackupConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create backup", slog.String("error", err.Error()))
		// Continue anyway, backup is not critical
	}

	if err := clientInstaller.ValidateConfig(ctx, currentConfig); err != nil {
		slog.ErrorContext(ctx, "invalid config format", slog.String("error", err.Error()))
		return nil, err
	}

	mcpServer := installer.NewKirhaMcpServer(config.ApiKey, config.Vertical)

	updatedConfig, err := clientInstaller.AddMcpServer(ctx, currentConfig, mcpServer)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add MCP server", slog.String("error", err.Error()))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				slog.ErrorContext(ctx, "failed to restore backup", slog.String("error", restoreErr.Error()))
			}
		}

		return nil, err
	}

	if err := clientInstaller.SaveConfig(ctx, updatedConfig); err != nil {
		slog.ErrorContext(ctx, "failed to save config", slog.String("error", err.Error()))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				slog.ErrorContext(ctx, "failed to restore backup", slog.String("error", restoreErr.Error()))
			}
		}

		return nil, err
	}

	savedConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load saved config for validation", slog.String("error", err.Error()))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				slog.ErrorContext(ctx, "failed to restore backup", slog.String("error", restoreErr.Error()))
			}
		}

		return nil, fmt.Errorf("%w: config validation failed", errors.ErrInstallationFailed)
	}

	if err := clientInstaller.ValidateConfig(ctx, savedConfig); err != nil {
		slog.ErrorContext(ctx, "saved config validation failed", slog.String("error", err.Error()))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				slog.ErrorContext(ctx, "failed to restore backup", slog.String("error", restoreErr.Error()))
			}
		}

		return nil, fmt.Errorf("%w: saved config invalid", errors.ErrInstallationFailed)
	}

	running, _ := clientInstaller.IsClientRunning(ctx)

	message := fmt.Sprintf("Successfully %s Kirha MCP server for %s", operation, config.Client)
	if running {
		message += ". Please restart the application to activate the MCP server."
	}

	slog.InfoContext(ctx, fmt.Sprintf("%s completed successfully", operation),
		slog.String("config_path", configPath),
		slog.String("backup_path", backupPath))

	return &installer.InstallResult{
		Success:    true,
		ConfigPath: configPath,
		BackupPath: backupPath,
		Message:    message,
	}, nil
}

func (a *Application) show(ctx context.Context, config *installer.Config) (*installer.ShowResult, error) {
	slog.InfoContext(ctx, "showing configuration",
		slog.String("client", string(config.Client)))

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get installer for client",
			slog.String("error", err.Error()),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get config path", slog.String("error", err.Error()))
		return nil, err
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load config", slog.String("error", err.Error()))
		return &installer.ShowResult{
			Success:    false,
			ConfigPath: configPath,
			HasServer:  false,
			Message:    fmt.Sprintf("Configuration file not found at %s", configPath),
		}, nil
	}

	var hasServer bool
	var serverConfig *installer.McpServer

	// If specific vertical is requested, check for that specific server
	if config.Vertical != "" {
		hasServer, err = clientInstaller.HasMcpServer(ctx, currentConfig, config.Vertical)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check if server exists", slog.String("error", err.Error()))
			return nil, err
		}

		if hasServer {
			serverConfig, err = clientInstaller.GetMcpServerConfig(ctx, currentConfig, config.Vertical)
			if err != nil {
				slog.ErrorContext(ctx, "failed to get server config", slog.String("error", err.Error()))
				return nil, err
			}
		}
	} else {
		// If no specific vertical requested, check if any Kirha servers exist
		hasCrypto, err := clientInstaller.HasMcpServer(ctx, currentConfig, installer.VerticalTypeCrypto)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check if crypto server exists", slog.String("error", err.Error()))
			return nil, err
		}

		// TODO : Add more verticals as needed
		hasServer = hasCrypto
	}

	// Handle vertical filtering by adding a new method to the interface
	var fullConfig string

	if config.Vertical != "" && hasServer {
		// Format only the specific vertical's server configuration
		fullConfig, err = clientInstaller.FormatSpecificServer(ctx, currentConfig, config.Vertical)
		if err != nil {
			slog.ErrorContext(ctx, "failed to format specific server config", slog.String("error", err.Error()))
			// Fall back to showing all servers
			fullConfig, err = clientInstaller.FormatConfig(ctx, currentConfig, config.OnlyKirha)
		}
	} else {
		fullConfig, err = clientInstaller.FormatConfig(ctx, currentConfig, config.OnlyKirha)
	}

	if err != nil {
		slog.ErrorContext(ctx, "failed to format config", slog.String("error", err.Error()))
		return nil, err
	}

	var message string
	if config.Vertical != "" {
		// Showing specific vertical
		if hasServer {
			message = fmt.Sprintf("MCP configuration for %s (vertical: %s):\n\n%s", config.Client, config.Vertical, fullConfig)
		} else {
			if fullConfig == "No MCP servers configured" {
				message = fmt.Sprintf("No MCP servers configured for %s", config.Client)
			} else {
				message = fmt.Sprintf("Kirha MCP server not found for %s %s vertical, but other servers are configured:\n\n%s", config.Client, config.Vertical, fullConfig)
			}
		}
	} else {
		// Showing all servers
		if fullConfig == "No MCP servers configured" {
			message = fmt.Sprintf("No MCP servers configured for %s", config.Client)
		} else {
			message = fmt.Sprintf("MCP configuration for %s:\n\n%s", config.Client, fullConfig)
		}
	}

	slog.InfoContext(ctx, "configuration displayed successfully",
		slog.String("config_path", configPath),
		slog.Bool("has_server", hasServer))

	return &installer.ShowResult{
		Success:      true,
		ConfigPath:   configPath,
		HasServer:    hasServer,
		ServerConfig: serverConfig,
		FullConfig:   fullConfig,
		Message:      message,
	}, nil
}

func (a *Application) Uninstall(ctx context.Context, config *installer.Config) (*installer.InstallResult, error) {
	slog.InfoContext(ctx, "starting uninstallation",
		slog.String("client", string(config.Client)),
		slog.Bool("dry_run", config.DryRun))

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get installer for client",
			slog.String("error", err.Error()),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	running, err := clientInstaller.IsClientRunning(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to check if client is running", slog.String("error", err.Error()))
	}
	if running && !config.DryRun {
		slog.WarnContext(ctx, "client is currently running",
			slog.String("client", string(config.Client)))
		return nil, errors.ErrClientRunning
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get config path", slog.String("error", err.Error()))
		return nil, err
	}

	if config.DryRun {
		slog.InfoContext(ctx, "dry run - would modify config",
			slog.String("path", configPath))
		return &installer.InstallResult{
			Success:    true,
			ConfigPath: configPath,
			Message:    fmt.Sprintf("Would remove Kirha MCP server from %s", configPath),
		}, nil
	}

	backupPath, err := clientInstaller.BackupConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create backup", slog.String("error", err.Error()))
		// Continue anyway, backup is not critical
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load config", slog.String("error", err.Error()))
		return nil, err
	}

	updatedConfig, err := clientInstaller.RemoveMcpServer(ctx, currentConfig, config.Vertical)
	if err != nil {
		slog.ErrorContext(ctx, "failed to remove MCP server", slog.String("error", err.Error()))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				slog.ErrorContext(ctx, "failed to restore backup", slog.String("error", restoreErr.Error()))
			}
		}

		return nil, err
	}

	if err := clientInstaller.SaveConfig(ctx, updatedConfig); err != nil {
		slog.ErrorContext(ctx, "failed to save config", slog.String("error", err.Error()))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				slog.ErrorContext(ctx, "failed to restore backup", slog.String("error", restoreErr.Error()))
			}
		}

		return nil, err
	}

	message := fmt.Sprintf("Successfully removed Kirha MCP server from %s", config.Client)
	if running {
		message += ". Please restart the application to apply changes."
	}

	slog.InfoContext(ctx, "uninstallation completed successfully",
		slog.String("config_path", configPath),
		slog.String("backup_path", backupPath))

	return &installer.InstallResult{
		Success:    true,
		ConfigPath: configPath,
		BackupPath: backupPath,
		Message:    message,
	}, nil
}

func (a *Application) validateApiKey(apiKey string) error {
	if apiKey == "" {
		return errors.ErrApiKeyRequired
	}

	if len(apiKey) < 8 {
		return errors.ErrApiKeyInvalid
	}

	if strings.TrimSpace(apiKey) != apiKey || strings.Contains(apiKey, " ") {
		return errors.ErrApiKeyInvalid
	}

	return nil
}
