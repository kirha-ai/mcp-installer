package docker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.kirha.ai/mcp-installer/internal/adapters/installers"
	"go.kirha.ai/mcp-installer/internal/core/domain/errors"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
	"go.kirha.ai/mcp-installer/pkg/security"
)

const (
	configFileName = "docker-compose.yml"
)

type DockerComposeConfig struct {
	Version  string                   `yaml:"version"`
	Services map[string]ServiceConfig `yaml:"services"`
	Networks map[string]NetworkConfig `yaml:"networks,omitempty"`
	Volumes  map[string]VolumeConfig  `yaml:"volumes,omitempty"`
}

type ServiceConfig struct {
	Image       string            `yaml:"image,omitempty"`
	Command     []string          `yaml:"command,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Restart     string            `yaml:"restart,omitempty"`
	Networks    []string          `yaml:"networks,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
}

type NetworkConfig struct {
	Driver string `yaml:"driver,omitempty"`
}

type VolumeConfig struct {
	Driver string `yaml:"driver,omitempty"`
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
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, configFileName), nil
}

func (i *Installer) LoadConfig(ctx context.Context) (interface{}, error) {
	path, err := i.GetConfigPath()
	if err != nil {
		return nil, err
	}

	// For Docker, we'll create a new docker-compose.yml if it doesn't exist
	if !i.FileExists(path) {
		slog.InfoContext(ctx, "docker-compose.yml not found, will create new one", slog.String("path", path))
		return &DockerComposeConfig{
			Version:  "3.8",
			Services: make(map[string]ServiceConfig),
			Networks: map[string]NetworkConfig{
				"mcp": {Driver: "bridge"},
			},
		}, nil
	}

	// For existing docker-compose files, we won't modify them
	// Instead, we'll create a separate docker-compose.mcp.yml
	return nil, fmt.Errorf("existing docker-compose.yml found, will create docker-compose.mcp.yml instead")
}

// AddMcpServer adds the MCP server to the configuration
func (i *Installer) AddMcpServer(ctx context.Context, config interface{}, server *installer.McpServer) (interface{}, error) {
	// For Docker, we create a complete docker-compose configuration
	// Use the server name from the McpServer parameter
	serviceName := fmt.Sprintf("%s-mcp", server.Name)
	dockerConfig := &DockerComposeConfig{
		Version: "3.8",
		Services: map[string]ServiceConfig{
			serviceName: {
				Image: "node:18-alpine",
				Command: []string{
					"sh", "-c",
					fmt.Sprintf("npx -y @%s/mcp-server", server.Name),
				},
				Environment: server.Environment,
				Restart:     "unless-stopped",
				Networks:    []string{"mcp"},
			},
		},
		Networks: map[string]NetworkConfig{
			"mcp": {Driver: "bridge"},
		},
	}

	slog.InfoContext(ctx, "created Docker Compose configuration for MCP server")

	return dockerConfig, nil
}

// RemoveMcpServer removes the MCP server from the configuration
func (i *Installer) RemoveMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (interface{}, error) {
	// For Docker, we'll just remove the generated file
	return nil, nil
}

// SaveConfig saves the configuration to disk
func (i *Installer) SaveConfig(ctx context.Context, config interface{}) error {
	if config == nil {
		// Remove operation - delete the file
		path, err := i.GetConfigPath()
		if err != nil {
			return err
		}

		// Check if we created a separate MCP file
		mcpPath := filepath.Join(filepath.Dir(path), "docker-compose.mcp.yml")
		if i.FileExists(mcpPath) {
			return os.Remove(mcpPath)
		}

		return nil
	}

	dockerConfig, ok := config.(*DockerComposeConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}

	path, err := i.GetConfigPath()
	if err != nil {
		return err
	}

	// If docker-compose.yml already exists, create docker-compose.mcp.yml instead
	if i.FileExists(path) {
		path = filepath.Join(filepath.Dir(path), "docker-compose.mcp.yml")
		slog.InfoContext(ctx, "existing docker-compose.yml found, creating docker-compose.mcp.yml",
			slog.String("path", path))
	}

	// Generate YAML content
	yamlContent := i.generateDockerComposeYAML(dockerConfig)

	return i.WriteFile(path, []byte(yamlContent))
}

// generateDockerComposeYAML generates YAML content for docker-compose
func (i *Installer) generateDockerComposeYAML(config *DockerComposeConfig) string {
	yaml := fmt.Sprintf("version: '%s'\n\n", config.Version)

	yaml += "services:\n"
	for name, service := range config.Services {
		yaml += fmt.Sprintf("  %s:\n", name)
		if service.Image != "" {
			yaml += fmt.Sprintf("    image: %s\n", service.Image)
		}
		if len(service.Command) > 0 {
			yaml += "    command:\n"
			for _, cmd := range service.Command {
				yaml += fmt.Sprintf("      - %s\n", cmd)
			}
		}
		if len(service.Environment) > 0 {
			yaml += "    environment:\n"
			for k, v := range service.Environment {
				yaml += fmt.Sprintf("      %s: %s\n", k, v)
			}
		}
		if service.Restart != "" {
			yaml += fmt.Sprintf("    restart: %s\n", service.Restart)
		}
		if len(service.Networks) > 0 {
			yaml += "    networks:\n"
			for _, net := range service.Networks {
				yaml += fmt.Sprintf("      - %s\n", net)
			}
		}
	}

	if len(config.Networks) > 0 {
		yaml += "\nnetworks:\n"
		for name, network := range config.Networks {
			yaml += fmt.Sprintf("  %s:\n", name)
			if network.Driver != "" {
				yaml += fmt.Sprintf("    driver: %s\n", network.Driver)
			}
		}
	}

	return yaml
}

// BackupConfig creates a backup of the current configuration
func (i *Installer) BackupConfig(ctx context.Context) (string, error) {
	path, err := i.GetConfigPath()
	if err != nil {
		return "", err
	}

	// Check both regular and MCP-specific files
	mcpPath := filepath.Join(filepath.Dir(path), "docker-compose.mcp.yml")
	if i.FileExists(mcpPath) {
		return i.CreateBackup(mcpPath)
	}

	if i.FileExists(path) {
		return i.CreateBackup(path)
	}

	slog.InfoContext(ctx, "no existing docker-compose files to backup")
	return "", nil
}

// RestoreConfig restores configuration from backup
func (i *Installer) RestoreConfig(ctx context.Context, backupPath string) error {
	if backupPath == "" {
		return nil
	}

	// Determine target path based on backup filename
	var targetPath string
	if filepath.Base(backupPath) == fmt.Sprintf("docker-compose.mcp.yml.backup_%s", filepath.Ext(backupPath)) {
		targetPath = filepath.Join(filepath.Dir(backupPath), "docker-compose.mcp.yml")
	} else {
		path, err := i.GetConfigPath()
		if err != nil {
			return err
		}
		targetPath = path
	}

	return i.RestoreBackup(backupPath, targetPath)
}

// ValidateConfig validates the configuration format
func (i *Installer) ValidateConfig(ctx context.Context, config interface{}) error {
	if config == nil {
		return nil // Valid for remove operation
	}

	_, ok := config.(*DockerComposeConfig)
	if !ok {
		return errors.ErrConfigInvalid
	}
	return nil
}

// IsClientRunning checks if Docker is currently running
func (i *Installer) IsClientRunning(ctx context.Context) (bool, error) {
	// Check if Docker daemon is running
	cmd := exec.CommandContext(ctx, "docker", "info")
	err := cmd.Run()
	if err != nil {
		slog.InfoContext(ctx, "Docker daemon not running or not installed")
		return false, nil
	}

	// Check if any MCP containers are running (check all verticals)
	cmd = exec.CommandContext(ctx, "docker", "ps", "--filter", "name=-mcp", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	return len(output) > 0, nil
}

// HasMcpServer checks if the MCP server exists in the configuration
func (i *Installer) HasMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (bool, error) {
	// Handle nil config (used for remove operations)
	if config == nil {
		return false, nil
	}

	dockerConfig, ok := config.(*DockerComposeConfig)
	if !ok {
		return false, errors.ErrConfigInvalid
	}

	// Use vertical-specific service name
	serviceName := fmt.Sprintf("%s-mcp", installer.GetServerName(vertical))
	_, exists := dockerConfig.Services[serviceName]
	return exists, nil
}

// GetMcpServerConfig returns the current MCP server configuration if it exists
func (i *Installer) GetMcpServerConfig(ctx context.Context, config interface{}, vertical installer.VerticalType) (*installer.McpServer, error) {
	// Handle nil config (used for remove operations)
	if config == nil {
		return nil, errors.ErrServerNotFound
	}

	dockerConfig, ok := config.(*DockerComposeConfig)
	if !ok {
		return nil, errors.ErrConfigInvalid
	}

	// Use vertical-specific service name
	serverName := installer.GetServerName(vertical)
	serviceName := fmt.Sprintf("%s-mcp", serverName)
	service, exists := dockerConfig.Services[serviceName]
	if !exists {
		return nil, errors.ErrServerNotFound
	}

	// Docker MCP server configuration
	mcpServer := &installer.McpServer{
		Name:        serverName,
		Command:     "npx",
		Args:        []string{"-y", fmt.Sprintf("@%s/mcp-server", serverName)},
		Environment: service.Environment,
	}

	return mcpServer, nil
}

// FormatConfig returns a human-readable representation of the configuration
func (i *Installer) FormatConfig(ctx context.Context, config interface{}, onlyKirha bool) (string, error) {
	// Handle nil config (used for remove operations)
	if config == nil {
		return "No MCP servers configured", nil
	}

	dockerConfig, ok := config.(*DockerComposeConfig)
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	if len(dockerConfig.Services) == 0 {
		return "No MCP servers configured", nil
	}

	// Separate services into Kirha and Other categories
	kirhaServices := make(map[string]ServiceConfig)
	otherServices := make(map[string]ServiceConfig)

	for name, service := range dockerConfig.Services {
		if strings.Contains(name, "kirha-") {
			kirhaServices[name] = service
		} else {
			otherServices[name] = service
		}
	}

	// If onlyKirha is true, only show Kirha services
	if onlyKirha {
		if len(kirhaServices) == 0 {
			return "No Kirha MCP servers configured", nil
		}
		return i.formatServiceSection("Kirha MCP Servers", kirhaServices), nil
	}

	// Format both sections
	var result string

	// Add Kirha section if any exist
	if len(kirhaServices) > 0 {
		result += i.formatServiceSection("Kirha MCP Servers", kirhaServices)
	}

	// Add Other section if any exist
	if len(otherServices) > 0 {
		if len(kirhaServices) > 0 {
			result += "\n"
		}
		result += i.formatServiceSection("Other MCP Servers", otherServices)
	}

	return result, nil
}

func (i *Installer) formatServiceSection(sectionTitle string, services map[string]ServiceConfig) string {
	var result string
	result += fmt.Sprintf("=== %s ===\n\n", sectionTitle)

	for name, service := range services {
		result += fmt.Sprintf("Service: %s\n", name)

		if service.Image != "" {
			result += fmt.Sprintf("  Image: %s\n", service.Image)
		}

		if len(service.Command) > 0 {
			result += fmt.Sprintf("  Command: %v\n", service.Command)
		}

		if len(service.Environment) > 0 {
			result += "  Environment:\n"
			for k, v := range service.Environment {
				if k == "KIRHA_API_KEY" {
					result += fmt.Sprintf("    %s: %s\n", k, security.MaskAPIKey(v))
				} else {
					result += fmt.Sprintf("    %s: %s\n", k, v)
				}
			}
		}

		if service.Restart != "" {
			result += fmt.Sprintf("  Restart: %s\n", service.Restart)
		}

		if len(service.Networks) > 0 {
			result += fmt.Sprintf("  Networks: %v\n", service.Networks)
		}

		result += "\n"
	}

	return result
}

func (i *Installer) FormatSpecificServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (string, error) {
	// Handle nil config
	if config == nil {
		return "", errors.ErrServerNotFound
	}

	dockerConfig, ok := config.(*DockerComposeConfig)
	if !ok {
		return "", errors.ErrConfigInvalid
	}

	serviceName := installer.GetServerName(vertical)
	serviceConfig, exists := dockerConfig.Services[serviceName]
	if !exists {
		return "", errors.ErrServerNotFound
	}

	// Create a map with only the specific service
	specificService := map[string]ServiceConfig{
		serviceName: serviceConfig,
	}

	// Format just this one service
	return i.formatServiceSection(fmt.Sprintf("Kirha MCP Server (%s)", vertical), specificService), nil
}
