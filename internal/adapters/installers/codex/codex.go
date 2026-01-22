package codex

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"go.kirha.ai/mcp-installer/internal/adapters/installers"
	"go.kirha.ai/mcp-installer/internal/core/domain/errors"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
	"go.kirha.ai/mcp-installer/pkg/security"
)

const (
	configFileName = "config.toml"
	configDir      = ".codex"
)

type CodexConfig struct {
	McpServers map[string]McpServerConfig `toml:"mcp_servers"`
}

type McpServerConfig struct {
	URL         string            `toml:"url"`
	HTTPHeaders map[string]string `toml:"http_headers,omitempty"`
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
		slog.InfoContext(ctx, "config file not found, creating new one", slog.String("path", path))
		return &CodexConfig{
			McpServers: make(map[string]McpServerConfig),
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrConfigReadFailed, err)
	}

	var config CodexConfig
	if _, err := toml.Decode(string(data), &config); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrConfigInvalid, err)
	}

	if config.McpServers == nil {
		config.McpServers = make(map[string]McpServerConfig)
	}

	return &config, nil
}

func (i *Installer) AddMcpServer(ctx context.Context, config interface{}, server *installer.McpServer) (interface{}, error) {
	codexConfig, ok := config.(*CodexConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	if _, exists := codexConfig.McpServers[server.Name]; exists {
		return nil, errors.ErrServerAlreadyExists
	}

	codexConfig.McpServers[server.Name] = McpServerConfig{
		URL:         server.URL,
		HTTPHeaders: server.Headers,
	}

	slog.InfoContext(ctx, "added MCP server to configuration",
		slog.String("server", server.Name))

	return codexConfig, nil
}

func (i *Installer) RemoveMcpServer(ctx context.Context, config interface{}) (interface{}, error) {
	codexConfig, ok := config.(*CodexConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	serverName := installer.ServerName
	if _, exists := codexConfig.McpServers[serverName]; !exists {
		return nil, errors.ErrServerNotFound
	}

	delete(codexConfig.McpServers, serverName)

	slog.InfoContext(ctx, "removed MCP server from configuration",
		slog.String("server", serverName))

	return codexConfig, nil
}

func (i *Installer) SaveConfig(ctx context.Context, config interface{}) error {
	codexConfig, ok := config.(*CodexConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}

	path, err := i.GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrConfigWriteFailed, err)
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(codexConfig); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrConfigWriteFailed, err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrConfigWriteFailed, err)
	}

	slog.InfoContext(ctx, "saved configuration", slog.String("path", path))
	return nil
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
	_, ok := config.(*CodexConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}
	return nil
}

func (i *Installer) IsClientRunning(ctx context.Context) (bool, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd := exec.CommandContext(ctx, "pgrep", "-f", "codex")
		err := cmd.Run()
		return err == nil, nil
	case "windows":
		cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq codex.exe")
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
	codexConfig, ok := config.(*CodexConfig)
	if !ok {
		return false, errors.ErrConfigInvalid
	}

	serverName := installer.ServerName
	_, exists := codexConfig.McpServers[serverName]
	return exists, nil
}

func (i *Installer) GetMcpServerConfig(ctx context.Context, config interface{}) (*installer.McpServer, error) {
	codexConfig, ok := config.(*CodexConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	serverName := installer.ServerName
	serverConfig, exists := codexConfig.McpServers[serverName]
	if !exists {
		return nil, errors.ErrServerNotFound
	}

	return &installer.McpServer{
		Name:    serverName,
		Type:    "http",
		URL:     serverConfig.URL,
		Headers: serverConfig.HTTPHeaders,
	}, nil
}

func (i *Installer) FormatConfig(ctx context.Context, config interface{}) (string, error) {
	codexConfig, ok := config.(*CodexConfig)
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	if len(codexConfig.McpServers) == 0 {
		return "No MCP servers configured", nil
	}

	kirhaServers := make(map[string]McpServerConfig)
	otherServers := make(map[string]McpServerConfig)

	for name, server := range codexConfig.McpServers {
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
		result += fmt.Sprintf("  URL: %s\n", server.URL)
		if len(server.HTTPHeaders) > 0 {
			result += "  Headers:\n"
			for k, v := range server.HTTPHeaders {
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
	codexConfig, ok := config.(*CodexConfig)
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	serverName := installer.ServerName
	serverConfig, exists := codexConfig.McpServers[serverName]
	if !exists {
		return "", errors.ErrServerNotFound
	}

	specificServer := map[string]McpServerConfig{
		serverName: serverConfig,
	}

	return i.formatServerSection("Kirha MCP Server", specificServer), nil
}
