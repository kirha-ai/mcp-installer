package claude

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"

	"github.com/kirha-ai/logger"
	"github.com/kirha-ai/mcp-installer/internal/adapters/installers"
	"github.com/kirha-ai/mcp-installer/internal/core/domain/errors"
	"github.com/kirha-ai/mcp-installer/internal/core/domain/installer"
	"github.com/kirha-ai/mcp-installer/pkg/security"
)

const (
	configFileName = "claude_desktop_config.json"
	appName        = "Claude"
	serverKey      = "mcpServers"
)

type ClaudeConfig struct {
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
		logger:        logger.New("claude_installer"),
	}
}

func (i *Installer) GetConfigPath() (string, error) {
	return i.GetPlatformConfigPath(appName, configFileName)
}

func (i *Installer) LoadConfig(ctx context.Context) (interface{}, error) {
	path, err := i.GetConfigPath()
	if err != nil {
		return nil, err
	}

	if !i.FileExists(path) {
		i.logger.InfoContext(ctx, "config file not found, creating new one", slog.String("path", path))
		return &ClaudeConfig{
			McpServers: make(map[string]McpServerConfig),
		}, nil
	}

	data, err := i.LoadJSONConfig(ctx, path)
	if err != nil {
		return nil, err
	}

	config := &ClaudeConfig{
		McpServers: make(map[string]McpServerConfig),
	}

	if servers, ok := data[serverKey].(map[string]interface{}); ok {
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
	claudeConfig, ok := config.(*ClaudeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	if _, exists := claudeConfig.McpServers[server.Name]; exists {
		return nil, errors.ErrServerAlreadyExists
	}

	claudeConfig.McpServers[server.Name] = McpServerConfig{
		Command: server.Command,
		Args:    server.Args,
		Env:     server.Environment,
	}

	i.logger.InfoContext(ctx, "added MCP server to configuration",
		slog.String("server", server.Name))

	return claudeConfig, nil
}

func (i *Installer) RemoveMcpServer(ctx context.Context, config interface{}) (interface{}, error) {
	claudeConfig, ok := config.(*ClaudeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	if _, exists := claudeConfig.McpServers[installer.ServerName]; !exists {
		return nil, errors.ErrServerNotFound
	}

	delete(claudeConfig.McpServers, "kirha")

	i.logger.InfoContext(ctx, "removed MCP server from configuration")

	return claudeConfig, nil
}

func (i *Installer) SaveConfig(ctx context.Context, config interface{}) error {
	claudeConfig, ok := config.(*ClaudeConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}

	path, err := i.GetConfigPath()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		serverKey: claudeConfig.McpServers,
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
	_, ok := config.(*ClaudeConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}
	return nil
}

func (i *Installer) IsClientRunning(ctx context.Context) (bool, error) {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.CommandContext(ctx, "pgrep", "-x", "Claude")
		err := cmd.Run()
		return err == nil, nil
	case "windows":
		cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq Claude.exe")
		output, err := cmd.Output()
		if err != nil {
			return false, nil
		}
		return len(output) > 0 && string(output) != "INFO: No tasks are running which match the specified criteria.", nil
	case "linux":
		cmd := exec.CommandContext(ctx, "pgrep", "-x", "claude")
		err := cmd.Run()
		return err == nil, nil
	default:
		return false, fmt.Errorf("%w: %s", errors.ErrPlatformNotSupported, runtime.GOOS)
	}
}

func (i *Installer) HasMcpServer(ctx context.Context, config interface{}) (bool, error) {
	claudeConfig, ok := config.(*ClaudeConfig)
	if !ok {
		return false, errors.ErrConfigInvalid
	}

	_, exists := claudeConfig.McpServers[installer.ServerName]
	return exists, nil
}

func (i *Installer) GetMcpServerConfig(ctx context.Context, config interface{}) (*installer.McpServer, error) {
	claudeConfig, ok := config.(*ClaudeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	serverConfig, exists := claudeConfig.McpServers[installer.ServerName]
	if !exists {
		return nil, errors.ErrServerNotFound
	}

	return &installer.McpServer{
		Name:        installer.ServerName,
		Command:     serverConfig.Command,
		Args:        serverConfig.Args,
		Environment: serverConfig.Env,
	}, nil
}

func (i *Installer) FormatConfig(ctx context.Context, config interface{}) (string, error) {
	claudeConfig, ok := config.(*ClaudeConfig)
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	if len(claudeConfig.McpServers) == 0 {
		return "No MCP servers configured", nil
	}

	var result string
	for name, server := range claudeConfig.McpServers {
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

	return result, nil
}
