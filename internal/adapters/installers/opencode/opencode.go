package opencode

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"go.kirha.ai/mcp-installer/internal/adapters/installers"
	"go.kirha.ai/mcp-installer/internal/core/domain/errors"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
	"go.kirha.ai/mcp-installer/pkg/security"
)

const (
	configFileName = "opencode.json"
	configDir      = "opencode"
	mcpKey         = "mcp"
)

type OpenCodeConfig struct {
	McpServers map[string]McpServerConfig `json:"mcp,omitempty"`
}

type McpServerConfig struct {
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Enabled bool              `json:"enabled"`
	Headers map[string]string `json:"headers,omitempty"`
}

type Installer struct {
	*installers.BaseInstaller
}

func New() *Installer {
	return &Installer{
		BaseInstaller: installers.NewBaseInstaller(),
	}
}

func (i *Installer) GetConfigPath() (string, error) {
	configBase, err := i.GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configBase, configDir, configFileName), nil
}

func (i *Installer) GetConfigDir() (string, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		home, err := i.GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config"), nil
	case "windows":
		return i.GetPlatformConfigPath("", "")
	default:
		return "", fmt.Errorf("%w: %s", errors.ErrPlatformNotSupported, runtime.GOOS)
	}
}

func (i *Installer) LoadConfig(ctx context.Context) (interface{}, error) {
	path, err := i.GetConfigPath()
	if err != nil {
		return nil, err
	}

	if !i.FileExists(path) {
		slog.InfoContext(ctx, "config file not found, creating new one", slog.String("path", path))
		return &OpenCodeConfig{
			McpServers: make(map[string]McpServerConfig),
		}, nil
	}

	data, err := i.LoadJSONConfig(ctx, path)
	if err != nil {
		return nil, err
	}

	config := &OpenCodeConfig{
		McpServers: make(map[string]McpServerConfig),
	}

	if servers, ok := data[mcpKey].(map[string]interface{}); ok {
		for name, serverData := range servers {
			if serverMap, ok := serverData.(map[string]interface{}); ok {
				mcpServer := McpServerConfig{}

				if serverType, ok := serverMap["type"].(string); ok {
					mcpServer.Type = serverType
				}

				if url, ok := serverMap["url"].(string); ok {
					mcpServer.URL = url
				}

				if enabled, ok := serverMap["enabled"].(bool); ok {
					mcpServer.Enabled = enabled
				}

				if headers, ok := serverMap["headers"].(map[string]interface{}); ok {
					mcpServer.Headers = make(map[string]string)
					for k, v := range headers {
						if vStr, ok := v.(string); ok {
							mcpServer.Headers[k] = vStr
						}
					}
				}

				config.McpServers[name] = mcpServer
			}
		}
	}

	return config, nil
}

func (i *Installer) AddMcpServer(ctx context.Context, config interface{}, server *installer.McpServer) (interface{}, error) {
	openCodeConfig, ok := config.(*OpenCodeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	if _, exists := openCodeConfig.McpServers[server.Name]; exists {
		return nil, errors.ErrServerAlreadyExists
	}

	// OpenCode uses "remote" type for HTTP MCP servers
	serverType := server.Type
	if serverType == "http" {
		serverType = "remote"
	}

	openCodeConfig.McpServers[server.Name] = McpServerConfig{
		Type:    serverType,
		URL:     server.URL,
		Enabled: true,
		Headers: server.Headers,
	}

	slog.InfoContext(ctx, "added MCP server to configuration",
		slog.String("server", server.Name))

	return openCodeConfig, nil
}

func (i *Installer) RemoveMcpServer(ctx context.Context, config interface{}) (interface{}, error) {
	openCodeConfig, ok := config.(*OpenCodeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	serverName := installer.ServerName
	if _, exists := openCodeConfig.McpServers[serverName]; !exists {
		return nil, errors.ErrServerNotFound
	}

	delete(openCodeConfig.McpServers, serverName)

	slog.InfoContext(ctx, "removed MCP server from configuration",
		slog.String("server", serverName))

	return openCodeConfig, nil
}

func (i *Installer) SaveConfig(ctx context.Context, config interface{}) error {
	openCodeConfig, ok := config.(*OpenCodeConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}

	path, err := i.GetConfigPath()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		mcpKey: openCodeConfig.McpServers,
	}

	if originalData, err := i.LoadJSONConfig(ctx, path); err == nil {
		for k, v := range originalData {
			if k != mcpKey {
				data[k] = v
			}
		}
	}

	return i.SaveJSONConfig(ctx, path, data)
}

func (i *Installer) BackupConfig(ctx context.Context) (string, error) {
	path, err := i.GetConfigPath()
	if err != nil {
		return "", err
	}

	if !i.FileExists(path) {
		slog.InfoContext(ctx, "no existing config to backup")
		return "", nil
	}

	return i.CreateBackup(path)
}

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

func (i *Installer) ValidateConfig(ctx context.Context, config interface{}) error {
	_, ok := config.(*OpenCodeConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}
	return nil
}

func (i *Installer) IsClientRunning(ctx context.Context) (bool, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd := exec.CommandContext(ctx, "pgrep", "-f", "opencode")
		err := cmd.Run()
		return err == nil, nil
	case "windows":
		cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq opencode.exe")
		output, err := cmd.Output()
		if err != nil {
			return false, nil
		}
		return len(output) > 0 && string(output) != "INFO: No tasks are running which match the specified criteria.", nil
	default:
		return false, fmt.Errorf("%w: %s", errors.ErrPlatformNotSupported, runtime.GOOS)
	}
}

func (i *Installer) HasMcpServer(ctx context.Context, config interface{}) (bool, error) {
	openCodeConfig, ok := config.(*OpenCodeConfig)
	if !ok {
		return false, errors.ErrConfigInvalid
	}

	serverName := installer.ServerName
	_, exists := openCodeConfig.McpServers[serverName]
	return exists, nil
}

func (i *Installer) GetMcpServerConfig(ctx context.Context, config interface{}) (*installer.McpServer, error) {
	openCodeConfig, ok := config.(*OpenCodeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	serverName := installer.ServerName
	serverConfig, exists := openCodeConfig.McpServers[serverName]
	if !exists {
		return nil, errors.ErrServerNotFound
	}

	return &installer.McpServer{
		Name:    serverName,
		Type:    serverConfig.Type,
		URL:     serverConfig.URL,
		Headers: serverConfig.Headers,
	}, nil
}

func (i *Installer) FormatConfig(ctx context.Context, config interface{}) (string, error) {
	openCodeConfig, ok := config.(*OpenCodeConfig)
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	if len(openCodeConfig.McpServers) == 0 {
		return "No MCP servers configured", nil
	}

	kirhaServers := make(map[string]McpServerConfig)
	otherServers := make(map[string]McpServerConfig)

	for name, server := range openCodeConfig.McpServers {
		if name == installer.ServerName || strings.HasPrefix(name, "kirha") {
			kirhaServers[name] = server
		} else {
			otherServers[name] = server
		}
	}

	var result string

	if len(kirhaServers) > 0 {
		result += i.formatServerSection("Kirha MCP Servers", kirhaServers)
	}

	if len(otherServers) > 0 {
		if len(kirhaServers) > 0 {
			result += "\n"
		}
		result += i.formatServerSection("Other MCP Servers", otherServers)
	}

	return result, nil
}

func (i *Installer) formatServerSection(sectionTitle string, servers map[string]McpServerConfig) string {
	var result string
	result += fmt.Sprintf("=== %s ===\n\n", sectionTitle)

	for name, server := range servers {
		result += fmt.Sprintf("Server: %s\n", name)
		result += fmt.Sprintf("  Type: %s\n", server.Type)
		result += fmt.Sprintf("  URL: %s\n", server.URL)
		if len(server.Headers) > 0 {
			result += "  Headers:\n"
			for k, v := range server.Headers {
				if k == "Authorization" {
					result += fmt.Sprintf("    %s: %s\n", k, security.MaskAPIKey(v))
				} else {
					result += fmt.Sprintf("    %s: %s\n", k, v)
				}
			}
		}
		result += "\n"
	}

	return result
}

func (i *Installer) FormatSpecificServer(ctx context.Context, config interface{}) (string, error) {
	openCodeConfig, ok := config.(*OpenCodeConfig)
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	serverName := installer.ServerName
	serverConfig, exists := openCodeConfig.McpServers[serverName]
	if !exists {
		return "", errors.ErrServerNotFound
	}

	specificServer := map[string]McpServerConfig{
		serverName: serverConfig,
	}

	return i.formatServerSection("Kirha MCP Server", specificServer), nil
}
