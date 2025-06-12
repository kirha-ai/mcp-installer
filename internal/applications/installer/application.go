package installer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/kirha-ai/logger"
	"github.com/kirha-ai/mcp-installer/internal/core/domain/errors"
	"github.com/kirha-ai/mcp-installer/internal/core/domain/installer"
	"github.com/kirha-ai/mcp-installer/internal/core/ports"
	"github.com/kirha-ai/mcp-installer/internal/core/ports/factories"
)

type Application struct {
	installerFactory factories.InstallerFactory
	logger           *slog.Logger
}

func New(installerFactory factories.InstallerFactory) *Application {
	return &Application{
		installerFactory: installerFactory,
		logger:           logger.New("installer_application"),
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
		a.logger.ErrorContext(ctx, "unknown operation requested",
			slog.String("operation", string(config.Operation)))
		return nil, errors.ErrUnknownOperation
	}
}

func (a *Application) install(ctx context.Context, config *installer.Config) (*installer.InstallResult, error) {
	a.logger.InfoContext(ctx, "starting installation",
		slog.String("client", string(config.Client)),
		slog.Bool("dry_run", config.DryRun))

	if err := a.validateApiKey(config.ApiKey); err != nil {
		return nil, err
	}

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get installer for client",
			logger.Error(err),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	running, err := clientInstaller.IsClientRunning(ctx)
	if err != nil {
		a.logger.WarnContext(ctx, "failed to check if client is running", logger.Error(err))
	}
	if running && !config.DryRun {
		a.logger.WarnContext(ctx, "client is currently running",
			slog.String("client", string(config.Client)))
		return nil, errors.ErrClientRunning
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get config path", logger.Error(err))
		return nil, err
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to load config", logger.Error(err))
		return nil, err
	}

	exists, err := clientInstaller.HasMcpServer(ctx, currentConfig)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to check if server exists", logger.Error(err))
		return nil, err
	}

	if exists {
		a.logger.ErrorContext(ctx, "MCP server already exists, use 'update' command to modify it")
		return nil, errors.ErrServerExistsUseUpdate
	}

	if config.DryRun {
		a.logger.InfoContext(ctx, "dry run - would install server",
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
	a.logger.InfoContext(ctx, "starting update",
		slog.String("client", string(config.Client)),
		slog.Bool("dry_run", config.DryRun))

	if err := a.validateApiKey(config.ApiKey); err != nil {
		return nil, err
	}

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get installer for client",
			logger.Error(err),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	running, err := clientInstaller.IsClientRunning(ctx)
	if err != nil {
		a.logger.WarnContext(ctx, "failed to check if client is running", logger.Error(err))
	}
	if running && !config.DryRun {
		a.logger.WarnContext(ctx, "client is currently running",
			slog.String("client", string(config.Client)))
		return nil, errors.ErrClientRunning
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get config path", logger.Error(err))
		return nil, err
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to load config", logger.Error(err))
		return nil, err
	}

	exists, err := clientInstaller.HasMcpServer(ctx, currentConfig)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to check if server exists", logger.Error(err))
		return nil, err
	}

	if !exists {
		a.logger.ErrorContext(ctx, "MCP server not found, use 'install' command to add it")
		return nil, errors.ErrServerNotFoundForUpdate
	}

	if config.DryRun {
		a.logger.InfoContext(ctx, "dry run - would update server",
			slog.String("path", configPath))
		return &installer.InstallResult{
			Success:    true,
			ConfigPath: configPath,
			Message:    fmt.Sprintf("Would update Kirha MCP server in %s", configPath),
		}, nil
	}

	configWithoutServer, err := clientInstaller.RemoveMcpServer(ctx, currentConfig)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to remove existing server", logger.Error(err))
		return nil, err
	}

	return a.performInstallOrUpdate(ctx, config, configWithoutServer, clientInstaller, "updated")
}

func (a *Application) remove(ctx context.Context, config *installer.Config) (*installer.InstallResult, error) {
	a.logger.InfoContext(ctx, "starting removal",
		slog.String("client", string(config.Client)),
		slog.Bool("dry_run", config.DryRun))

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get installer for client",
			logger.Error(err),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	running, err := clientInstaller.IsClientRunning(ctx)
	if err != nil {
		a.logger.WarnContext(ctx, "failed to check if client is running", logger.Error(err))
	}
	if running && !config.DryRun {
		a.logger.WarnContext(ctx, "client is currently running",
			slog.String("client", string(config.Client)))
		return nil, errors.ErrClientRunning
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get config path", logger.Error(err))
		return nil, err
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to load config", logger.Error(err))
		return nil, err
	}

	exists, err := clientInstaller.HasMcpServer(ctx, currentConfig)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to check if server exists", logger.Error(err))
		return nil, err
	}

	if !exists {
		a.logger.ErrorContext(ctx, "MCP server not found, nothing to remove")
		return nil, errors.ErrServerNotFoundForRemove
	}

	if config.DryRun {
		a.logger.InfoContext(ctx, "dry run - would remove server",
			slog.String("path", configPath))
		return &installer.InstallResult{
			Success:    true,
			ConfigPath: configPath,
			Message:    fmt.Sprintf("Would remove Kirha MCP server from %s", configPath),
		}, nil
	}

	backupPath, err := clientInstaller.BackupConfig(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to create backup", logger.Error(err))
	}

	updatedConfig, err := clientInstaller.RemoveMcpServer(ctx, currentConfig)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to remove MCP server", logger.Error(err))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				a.logger.ErrorContext(ctx, "failed to restore backup", logger.Error(restoreErr))
			}
		}

		return nil, err
	}

	if err := clientInstaller.SaveConfig(ctx, updatedConfig); err != nil {
		a.logger.ErrorContext(ctx, "failed to save config", logger.Error(err))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				a.logger.ErrorContext(ctx, "failed to restore backup", logger.Error(restoreErr))
			}
		}

		return nil, err
	}

	message := fmt.Sprintf("Successfully removed Kirha MCP server from %s", config.Client)
	if running {
		message += ". Please restart the application to apply changes."
	}

	a.logger.InfoContext(ctx, "removal completed successfully",
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
		a.logger.ErrorContext(ctx, "failed to get config path", logger.Error(err))
		return nil, err
	}

	backupPath, err := clientInstaller.BackupConfig(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to create backup", logger.Error(err))
		// Continue anyway, backup is not critical
	}

	if err := clientInstaller.ValidateConfig(ctx, currentConfig); err != nil {
		a.logger.ErrorContext(ctx, "invalid config format", logger.Error(err))
		return nil, err
	}

	mcpServer := installer.NewKirhaMcpServer(config.ApiKey)

	updatedConfig, err := clientInstaller.AddMcpServer(ctx, currentConfig, mcpServer)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to add MCP server", logger.Error(err))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				a.logger.ErrorContext(ctx, "failed to restore backup", logger.Error(restoreErr))
			}
		}

		return nil, err
	}

	if err := clientInstaller.SaveConfig(ctx, updatedConfig); err != nil {
		a.logger.ErrorContext(ctx, "failed to save config", logger.Error(err))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				a.logger.ErrorContext(ctx, "failed to restore backup", logger.Error(restoreErr))
			}
		}

		return nil, err
	}

	savedConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to load saved config for validation", logger.Error(err))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				a.logger.ErrorContext(ctx, "failed to restore backup", logger.Error(restoreErr))
			}
		}

		return nil, fmt.Errorf("%w: config validation failed", errors.ErrInstallationFailed)
	}

	if err := clientInstaller.ValidateConfig(ctx, savedConfig); err != nil {
		a.logger.ErrorContext(ctx, "saved config validation failed", logger.Error(err))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				a.logger.ErrorContext(ctx, "failed to restore backup", logger.Error(restoreErr))
			}
		}

		return nil, fmt.Errorf("%w: saved config invalid", errors.ErrInstallationFailed)
	}

	running, _ := clientInstaller.IsClientRunning(ctx)

	message := fmt.Sprintf("Successfully %s Kirha MCP server for %s", operation, config.Client)
	if running {
		message += ". Please restart the application to activate the MCP server."
	}

	a.logger.InfoContext(ctx, fmt.Sprintf("%s completed successfully", operation),
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
	a.logger.InfoContext(ctx, "showing configuration",
		slog.String("client", string(config.Client)))

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get installer for client",
			logger.Error(err),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get config path", logger.Error(err))
		return nil, err
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to load config", logger.Error(err))
		return &installer.ShowResult{
			Success:    false,
			ConfigPath: configPath,
			HasServer:  false,
			Message:    fmt.Sprintf("Configuration file not found at %s", configPath),
		}, nil
	}

	hasServer, err := clientInstaller.HasMcpServer(ctx, currentConfig)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to check if server exists", logger.Error(err))
		return nil, err
	}

	var serverConfig *installer.McpServer
	var message string
	var fullConfig string

	if hasServer {
		serverConfig, err = clientInstaller.GetMcpServerConfig(ctx, currentConfig)
		if err != nil {
			a.logger.ErrorContext(ctx, "failed to get server config", logger.Error(err))
			return nil, err
		}

		fullConfig, err = clientInstaller.FormatConfig(ctx, currentConfig)
		if err != nil {
			a.logger.ErrorContext(ctx, "failed to format config", logger.Error(err))
			return nil, err
		}

		message = fmt.Sprintf("MCP configuration for %s:\n\n%s", config.Client, fullConfig)
	} else {
		fullConfig, err = clientInstaller.FormatConfig(ctx, currentConfig)
		if err != nil {
			a.logger.ErrorContext(ctx, "failed to format config", logger.Error(err))
			return nil, err
		}

		if fullConfig == "No MCP servers configured" {
			message = fmt.Sprintf("No MCP servers configured for %s", config.Client)
		} else {
			message = fmt.Sprintf("Kirha MCP server not found for %s, but other servers are configured:\n\n%s", config.Client, fullConfig)
		}
	}

	a.logger.InfoContext(ctx, "configuration displayed successfully",
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
	a.logger.InfoContext(ctx, "starting uninstallation",
		slog.String("client", string(config.Client)),
		slog.Bool("dry_run", config.DryRun))

	clientInstaller, err := a.installerFactory.GetInstaller(ctx, config.Client)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get installer for client",
			logger.Error(err),
			slog.String("client", string(config.Client)))
		return nil, err
	}

	running, err := clientInstaller.IsClientRunning(ctx)
	if err != nil {
		a.logger.WarnContext(ctx, "failed to check if client is running", logger.Error(err))
	}
	if running && !config.DryRun {
		a.logger.WarnContext(ctx, "client is currently running",
			slog.String("client", string(config.Client)))
		return nil, errors.ErrClientRunning
	}

	configPath, err := clientInstaller.GetConfigPath()
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get config path", logger.Error(err))
		return nil, err
	}

	if config.DryRun {
		a.logger.InfoContext(ctx, "dry run - would modify config",
			slog.String("path", configPath))
		return &installer.InstallResult{
			Success:    true,
			ConfigPath: configPath,
			Message:    fmt.Sprintf("Would remove Kirha MCP server from %s", configPath),
		}, nil
	}

	backupPath, err := clientInstaller.BackupConfig(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to create backup", logger.Error(err))
		// Continue anyway, backup is not critical
	}

	currentConfig, err := clientInstaller.LoadConfig(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to load config", logger.Error(err))
		return nil, err
	}

	updatedConfig, err := clientInstaller.RemoveMcpServer(ctx, currentConfig)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to remove MCP server", logger.Error(err))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				a.logger.ErrorContext(ctx, "failed to restore backup", logger.Error(restoreErr))
			}
		}

		return nil, err
	}

	if err := clientInstaller.SaveConfig(ctx, updatedConfig); err != nil {
		a.logger.ErrorContext(ctx, "failed to save config", logger.Error(err))

		if backupPath != "" {
			if restoreErr := clientInstaller.RestoreConfig(ctx, backupPath); restoreErr != nil {
				a.logger.ErrorContext(ctx, "failed to restore backup", logger.Error(restoreErr))
			}
		}

		return nil, err
	}

	message := fmt.Sprintf("Successfully removed Kirha MCP server from %s", config.Client)
	if running {
		message += ". Please restart the application to apply changes."
	}

	a.logger.InfoContext(ctx, "uninstallation completed successfully",
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

	if strings.TrimSpace(apiKey) != apiKey || strings.Contains(apiKey, " ") {
		return errors.ErrApiKeyInvalid
	}

	return nil
}
