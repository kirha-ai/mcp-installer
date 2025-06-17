package vscode

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"

	"go.kirha.ai/mcp-installer/internal/adapters/installers"
	"go.kirha.ai/mcp-installer/internal/core/domain/errors"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
	"go.kirha.ai/mcp-installer/pkg/security"
)

const (
	configFileName = "settings.json"
	appName        = "Code"
	mcpKey         = "mcp.servers"
)

type Installer struct {
	*installers.BaseInstaller
}

func New() *Installer {
	return &Installer{
		BaseInstaller: installers.NewBaseInstaller(),
	}
}

func (i *Installer) GetConfigPath() (string, error) {
	return i.GetPlatformConfigPath(fmt.Sprintf("%s/User", appName), configFileName)
}

func (i *Installer) LoadConfig(ctx context.Context) (interface{}, error) {
	path, err := i.GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return empty config
	if !i.FileExists(path) {
		slog.InfoContext(ctx, "config file not found, creating new one", slog.String("path", path))
		return make(map[string]interface{}), nil
	}

	return i.LoadJSONConfig(ctx, path)
}

// AddMcpServer adds the MCP server to the configuration
func (i *Installer) AddMcpServer(ctx context.Context, config interface{}, server *installer.McpServer) (interface{}, error) {
	settings, ok := config.(map[string]interface{})
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	// Get or create mcp.servers section
	mcpServers, ok := settings[mcpKey].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
	}

	// Check if server already exists
	if _, exists := mcpServers[server.Name]; exists {
		return nil, errors.ErrServerAlreadyExists
	}

	// Add server configuration
	mcpServers[server.Name] = map[string]interface{}{
		"command": server.Command,
		"args":    server.Args,
		"env":     server.Environment,
	}

	settings[mcpKey] = mcpServers

	slog.InfoContext(ctx, "added MCP server to configuration",
		slog.String("server", server.Name))

	return settings, nil
}

// RemoveMcpServer removes the MCP server from the configuration
func (i *Installer) RemoveMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (interface{}, error) {
	settings, ok := config.(map[string]interface{})
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	// Get mcp.servers section
	mcpServers, ok := settings[mcpKey].(map[string]interface{})
	if !ok {
		return nil, errors.ErrServerNotFound
	}

	// Check if server exists
	serverName := installer.GetServerName(vertical)
	if _, exists := mcpServers[serverName]; !exists {
		return nil, errors.ErrServerNotFound
	}

	// Remove server
	delete(mcpServers, serverName)

	// Remove the entire mcp.servers key if it's empty
	if len(mcpServers) == 0 {
		delete(settings, mcpKey)
	} else {
		settings[mcpKey] = mcpServers
	}

	slog.InfoContext(ctx, "removed MCP server from configuration",
		slog.String("server", serverName))

	return settings, nil
}

// SaveConfig saves the configuration to disk
func (i *Installer) SaveConfig(ctx context.Context, config interface{}) error {
	path, err := i.GetConfigPath()
	if err != nil {
		return err
	}

	return i.SaveJSONConfig(ctx, path, config)
}

// BackupConfig creates a backup of the current configuration
func (i *Installer) BackupConfig(ctx context.Context) (string, error) {
	path, err := i.GetConfigPath()
	if err != nil {
		return "", err
	}

	// If config doesn't exist, no need to backup
	if !i.FileExists(path) {
		slog.InfoContext(ctx, "no existing config to backup")
		return "", nil
	}

	return i.CreateBackup(path)
}

// RestoreConfig restores configuration from backup
func (i *Installer) RestoreConfig(ctx context.Context, backupPath string) error {
	if backupPath == "" {
		return nil
	}

	path, err := i.GetConfigPath()
	if err != nil {
		return err
	}

	return i.RestoreBackup(backupPath, path)
}

// ValidateConfig validates the configuration format
func (i *Installer) ValidateConfig(ctx context.Context, config interface{}) error {
	_, ok := config.(map[string]interface{})
	if !ok {
		return errors.ErrConfigInvalid
	}
	return nil
}

// IsClientRunning checks if VS Code is currently running
func (i *Installer) IsClientRunning(ctx context.Context) (bool, error) {
	switch runtime.GOOS {
	case "darwin":
		// Check both "Code" and "Visual Studio Code"
		cmd := exec.CommandContext(ctx, "pgrep", "-f", "Visual Studio Code|Code Helper")
		err := cmd.Run()
		return err == nil, nil
	case "windows":
		cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq Code.exe")
		output, err := cmd.Output()
		if err != nil {
			return false, nil
		}
		return len(output) > 0 && string(output) != "INFO: No tasks are running which match the specified criteria.", nil
	case "linux":
		cmd := exec.CommandContext(ctx, "pgrep", "-x", "code")
		err := cmd.Run()
		return err == nil, nil
	default:
		return false, fmt.Errorf("%w: %s", errors.ErrPlatformNotSupported, runtime.GOOS)
	}
}

// HasMcpServer checks if the MCP server exists in the configuration
func (i *Installer) HasMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (bool, error) {
	settings, ok := config.(map[string]interface{})
	if !ok {
		return false, errors.ErrConfigInvalid
	}

	// Get mcp.servers section
	mcpServers, ok := settings[mcpKey].(map[string]interface{})
	if !ok {
		return false, nil
	}

	serverName := installer.GetServerName(vertical)
	_, exists := mcpServers[serverName]
	return exists, nil
}

// GetMcpServerConfig returns the current MCP server configuration if it exists
func (i *Installer) GetMcpServerConfig(ctx context.Context, config interface{}, vertical installer.VerticalType) (*installer.McpServer, error) {
	settings, ok := config.(map[string]interface{})
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	// Get mcp.servers section
	mcpServers, ok := settings[mcpKey].(map[string]interface{})
	if !ok {
		return nil, errors.ErrServerNotFound
	}

	serverName := installer.GetServerName(vertical)
	serverConfig, exists := mcpServers[serverName]
	if !exists {
		return nil, errors.ErrServerNotFound
	}

	serverMap, ok := serverConfig.(map[string]interface{})
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	mcpServer := &installer.McpServer{
		Name: serverName,
	}

	if cmd, ok := serverMap["command"].(string); ok {
		mcpServer.Command = cmd
	}

	if args, ok := serverMap["args"].([]interface{}); ok {
		mcpServer.Args = make([]string, len(args))
		for j, arg := range args {
			if argStr, ok := arg.(string); ok {
				mcpServer.Args[j] = argStr
			}
		}
	}

	if env, ok := serverMap["env"].(map[string]interface{}); ok {
		mcpServer.Environment = make(map[string]string)
		for k, v := range env {
			if vStr, ok := v.(string); ok {
				mcpServer.Environment[k] = vStr
			}
		}
	}

	return mcpServer, nil
}

// FormatConfig returns a human-readable representation of the configuration
func (i *Installer) FormatConfig(ctx context.Context, config interface{}, onlyKirha bool) (string, error) {
	settings, ok := config.(map[string]interface{})
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	// Get mcp.servers section
	mcpServers, ok := settings[mcpKey].(map[string]interface{})
	if !ok || len(mcpServers) == 0 {
		return "No MCP servers configured", nil
	}

	// Separate servers into Kirha and Other categories
	kirhaServers := make(map[string]interface{})
	otherServers := make(map[string]interface{})

	for name, serverConfig := range mcpServers {
		if strings.HasPrefix(name, "kirha-") {
			kirhaServers[name] = serverConfig
		} else {
			otherServers[name] = serverConfig
		}
	}

	// If onlyKirha is true, only show Kirha servers
	if onlyKirha {
		if len(kirhaServers) == 0 {
			return "No Kirha MCP servers configured", nil
		}
		return i.formatServerSection("Kirha MCP Servers", kirhaServers), nil
	}

	// Format both sections
	var result string

	// Add Kirha section if any exist
	if len(kirhaServers) > 0 {
		result += i.formatServerSection("Kirha MCP Servers", kirhaServers)
	}

	// Add Other section if any exist
	if len(otherServers) > 0 {
		if len(kirhaServers) > 0 {
			result += "\n"
		}
		result += i.formatServerSection("Other MCP Servers", otherServers)
	}

	return result, nil
}

func (i *Installer) formatServerSection(sectionTitle string, servers map[string]interface{}) string {
	var result string
	result += fmt.Sprintf("=== %s ===\n\n", sectionTitle)

	for name, serverConfig := range servers {
		serverMap, ok := serverConfig.(map[string]interface{})
		if !ok {
			continue
		}

		result += fmt.Sprintf("Server: %s\n", name)

		if cmd, ok := serverMap["command"].(string); ok {
			result += fmt.Sprintf("  Command: %s\n", cmd)
		}

		if args, ok := serverMap["args"].([]interface{}); ok {
			argStrs := make([]string, len(args))
			for j, arg := range args {
				if argStr, ok := arg.(string); ok {
					argStrs[j] = argStr
				}
			}
			result += fmt.Sprintf("  Args: %v\n", argStrs)
		}

		if env, ok := serverMap["env"].(map[string]interface{}); ok && len(env) > 0 {
			result += "  Environment:\n"
			for k, v := range env {
				if vStr, ok := v.(string); ok {
					if k == "KIRHA_API_KEY" {
						result += fmt.Sprintf("    %s: %s\n", k, security.MaskAPIKey(vStr))
					} else {
						result += fmt.Sprintf("    %s: %s\n", k, vStr)
					}
				}
			}
		}
		result += "\n"
	}

	return result
}

func (i *Installer) FormatSpecificServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (string, error) {
	settings, ok := config.(map[string]interface{})
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	// Get mcp.servers section
	mcpServers, ok := settings[mcpKey].(map[string]interface{})
	if !ok {
		return "", errors.ErrServerNotFound
	}

	serverName := installer.GetServerName(vertical)
	serverConfig, exists := mcpServers[serverName]
	if !exists {
		return "", errors.ErrServerNotFound
	}

	// Create a map with only the specific server
	specificServer := map[string]interface{}{
		serverName: serverConfig,
	}

	// Format just this one server
	return i.formatServerSection(fmt.Sprintf("Kirha MCP Server (%s)", vertical), specificServer), nil
}
