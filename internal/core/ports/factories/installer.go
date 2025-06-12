package factories

import (
	"context"
	
	"github.com/kirha-ai/mcp-installer/internal/core/domain/installer"
	"github.com/kirha-ai/mcp-installer/internal/core/ports"
)

type InstallerFactory interface {
	GetInstaller(ctx context.Context, clientType installer.ClientType) (ports.Installer, error)
}