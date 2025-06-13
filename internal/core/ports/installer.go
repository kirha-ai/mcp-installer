package ports

import (
	"context"

	"github.com/kirha-ai/mcp-installer/internal/core/domain/installer"
)

type Installer interface {
	GetConfigPath() (string, error)
	LoadConfig(ctx context.Context) (interface{}, error)
	AddMcpServer(ctx context.Context, config interface{}, server *installer.McpServer) (interface{}, error)
	RemoveMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (interface{}, error)
	SaveConfig(ctx context.Context, config interface{}) error
	BackupConfig(ctx context.Context) (string, error)
	RestoreConfig(ctx context.Context, backupPath string) error
	ValidateConfig(ctx context.Context, config interface{}) error
	IsClientRunning(ctx context.Context) (bool, error)
	HasMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (bool, error)
	GetMcpServerConfig(ctx context.Context, config interface{}, vertical installer.VerticalType) (*installer.McpServer, error)
	FormatConfig(ctx context.Context, config interface{}, onlyKirha bool) (string, error)
	FormatSpecificServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (string, error)
}

type ConfigManager interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte) error
	FileExists(path string) bool
	CreateBackup(path string) (string, error)
	RestoreBackup(backupPath, targetPath string) error
}
