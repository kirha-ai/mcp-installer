package errors

import "errors"

var (
	ErrConfigNotFound      = errors.New("configuration file not found")
	ErrConfigInvalid       = errors.New("invalid configuration format")
	ErrConfigReadFailed    = errors.New("failed to read configuration")
	ErrConfigWriteFailed   = errors.New("failed to write configuration")
	ErrConfigBackupFailed  = errors.New("failed to backup configuration")
	ErrConfigRestoreFailed = errors.New("failed to restore configuration")

	ErrClientNotSupported = errors.New("client not supported")
	ErrClientRunning      = errors.New("client is currently running, please close it before installing")

	ErrApiKeyRequired = errors.New("API key is required")
	ErrApiKeyInvalid  = errors.New("invalid API key format")

	ErrInstallationFailed   = errors.New("installation failed")
	ErrUninstallationFailed = errors.New("uninstallation failed")
	ErrUpdateFailed         = errors.New("update failed")
	ErrServerAlreadyExists  = errors.New("MCP server already exists in configuration")
	ErrServerNotFound       = errors.New("MCP server not found in configuration")

	ErrServerExistsUseUpdate   = errors.New("MCP server already exists, use 'update' command to modify it")
	ErrServerNotFoundForUpdate = errors.New("MCP server not found, use 'install' command to add it")
	ErrServerNotFoundForRemove = errors.New("MCP server not found, nothing to remove")

	ErrPathNotFound     = errors.New("path not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrBackupExists     = errors.New("backup file already exists")

	ErrPlatformNotSupported = errors.New("platform not supported")

	ErrUnknownOperation  = errors.New("unknown operation")
	ErrUnsupportedClient = errors.New("unsupported client")
)
