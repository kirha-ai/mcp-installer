package factories

import (
	"context"

	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
	"go.kirha.ai/mcp-installer/internal/core/ports"
)

type InstallerFactory interface {
	GetInstaller(ctx context.Context, clientType installer.ClientType) (ports.Installer, error)
}
