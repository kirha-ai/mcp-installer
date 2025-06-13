package claudecode

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kirha-ai/logger"
	"github.com/kirha-ai/mcp-installer/internal/adapters/installers"
	"github.com/kirha-ai/mcp-installer/internal/core/domain/errors"
	"github.com/kirha-ai/mcp-installer/internal/core/domain/installer"
	"github.com/kirha-ai/mcp-installer/pkg/security"
)

const (
	configFileName = "config.json"
	configDir      = ".claude-code"
	mcpKey         = "mcpServers"
)

type ClaudeCodeConfig struct {
	McpServers map[string]McpServerConfig `json:"mcpServers,omitempty"`
}

type McpServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type Installer struct {
	*installers.BaseInstaller
	logger *slog.Logger
}

func New() *Installer {
	return &Installer{
		BaseInstaller: installers.NewBaseInstaller(),
		logger:        logger.New("claude_code_installer"),
	}
}

func (i *Installer) GetConfigPath() (string, error) {
	home, err := i.GetHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, configDir, configFileName), nil
}

func (i *Installer) LoadConfig(ctx context.Context) (interface{}, error) {
	path, err := i.GetConfigPath()
	if err != nil {
		return nil, err
	}

	if !i.FileExists(path) {
		i.logger.InfoContext(ctx, "config file not found, creating new one", slog.String("path", path))
		return &ClaudeCodeConfig{
			McpServers: make(map[string]McpServerConfig),
		}, nil
	}

	data, err := i.LoadJSONConfig(ctx, path)
	if err != nil {
		return nil, err
	}

	config := &ClaudeCodeConfig{
		McpServers: make(map[string]McpServerConfig),
	}

	if servers, ok := data[mcpKey].(map[string]interface{}); ok {
		for name, serverData := range servers {
			if serverMap, ok := serverData.(map[string]interface{}); ok {
				mcpServer := McpServerConfig{}

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
					mcpServer.Env = make(map[string]string)
					for k, v := range env {
						if vStr, ok := v.(string); ok {
							mcpServer.Env[k] = vStr
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
	claudeCodeConfig, ok := config.(*ClaudeCodeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	if _, exists := claudeCodeConfig.McpServers[server.Name]; exists {
		return nil, errors.ErrServerAlreadyExists
	}

	claudeCodeConfig.McpServers[server.Name] = McpServerConfig{
		Command: server.Command,
		Args:    server.Args,
		Env:     server.Environment,
	}

	i.logger.InfoContext(ctx, "added MCP server to configuration",
		slog.String("server", server.Name))

	return claudeCodeConfig, nil
}

func (i *Installer) RemoveMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (interface{}, error) {
	claudeCodeConfig, ok := config.(*ClaudeCodeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	serverName := installer.GetServerName(vertical)
	if _, exists := claudeCodeConfig.McpServers[serverName]; !exists {
		return nil, errors.ErrServerNotFound
	}

	delete(claudeCodeConfig.McpServers, serverName)

	i.logger.InfoContext(ctx, "removed MCP server from configuration",
		slog.String("server", serverName))

	return claudeCodeConfig, nil
}

func (i *Installer) SaveConfig(ctx context.Context, config interface{}) error {
	claudeCodeConfig, ok := config.(*ClaudeCodeConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}

	path, err := i.GetConfigPath()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		mcpKey: claudeCodeConfig.McpServers,
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
		i.logger.InfoContext(ctx, "no existing config to backup")
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
	_, ok := config.(*ClaudeCodeConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}
	return nil
}

func (i *Installer) IsClientRunning(ctx context.Context) (bool, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd := exec.CommandContext(ctx, "pgrep", "-f", "claude.*code|claude-code")
		err := cmd.Run()
		return err == nil, nil
	case "windows":
		cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq claude.exe")
		output, err := cmd.Output()
		if err != nil {
			return false, nil
		}
		return len(output) > 0 && string(output) != "INFO: No tasks are running which match the specified criteria.", nil
	default:
		return false, fmt.Errorf("%w: %s", errors.ErrPlatformNotSupported, runtime.GOOS)
	}
}

func (i *Installer) HasMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (bool, error) {
	claudeCodeConfig, ok := config.(*ClaudeCodeConfig)
	if !ok {
		return false, errors.ErrConfigInvalid
	}

	serverName := installer.GetServerName(vertical)
	_, exists := claudeCodeConfig.McpServers[serverName]
	return exists, nil
}

func (i *Installer) GetMcpServerConfig(ctx context.Context, config interface{}, vertical installer.VerticalType) (*installer.McpServer, error) {
	claudeCodeConfig, ok := config.(*ClaudeCodeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	serverName := installer.GetServerName(vertical)
	serverConfig, exists := claudeCodeConfig.McpServers[serverName]
	if !exists {
		return nil, errors.ErrServerNotFound
	}

	return &installer.McpServer{
		Name:        serverName,
		Command:     serverConfig.Command,
		Args:        serverConfig.Args,
		Environment: serverConfig.Env,
	}, nil
}

func (i *Installer) FormatConfig(ctx context.Context, config interface{}, onlyKirha bool) (string, error) {
	claudeCodeConfig, ok := config.(*ClaudeCodeConfig)
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	if len(claudeCodeConfig.McpServers) == 0 {
		return "No MCP servers configured", nil
	}

	// Separate servers into Kirha and Other categories
	kirhaServers := make(map[string]McpServerConfig)
	otherServers := make(map[string]McpServerConfig)

	for name, server := range claudeCodeConfig.McpServers {
		if strings.HasPrefix(name, "kirha-") {
			kirhaServers[name] = server
		} else {
			otherServers[name] = server
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

func (i *Installer) formatServerSection(sectionTitle string, servers map[string]McpServerConfig) string {
	var result string
	result += fmt.Sprintf("=== %s ===\n\n", sectionTitle)

	for name, server := range servers {
		result += fmt.Sprintf("Server: %s\n", name)
		result += fmt.Sprintf("  Command: %s\n", server.Command)
		result += fmt.Sprintf("  Args: %v\n", server.Args)
		if len(server.Env) > 0 {
			result += "  Environment:\n"
			for k, v := range server.Env {
				if k == "KIRHA_API_KEY" {
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

func (i *Installer) FormatSpecificServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (string, error) {
	claudeCodeConfig, ok := config.(*ClaudeCodeConfig)
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	serverName := installer.GetServerName(vertical)
	serverConfig, exists := claudeCodeConfig.McpServers[serverName]
	if !exists {
		return "", errors.ErrServerNotFound
	}

	// Create a map with only the specific server
	specificServer := map[string]McpServerConfig{
		serverName: serverConfig,
	}

	// Format just this one server
	return i.formatServerSection(fmt.Sprintf("Kirha MCP Server (%s)", vertical), specificServer), nil
}
