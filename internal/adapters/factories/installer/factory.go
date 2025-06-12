package installerfactory

import (
	"context"
	
	"github.com/kirha-ai/mcp-installer/internal/adapters/installers/claude"
	claudecode "github.com/kirha-ai/mcp-installer/internal/adapters/installers/claude-code"
	"github.com/kirha-ai/mcp-installer/internal/adapters/installers/cursor"
	"github.com/kirha-ai/mcp-installer/internal/adapters/installers/docker"
	"github.com/kirha-ai/mcp-installer/internal/adapters/installers/vscode"
	"github.com/kirha-ai/mcp-installer/internal/core/domain/errors"
	"github.com/kirha-ai/mcp-installer/internal/core/domain/installer"
	"github.com/kirha-ai/mcp-installer/internal/core/ports"
	"github.com/kirha-ai/mcp-installer/internal/core/ports/factories"
)

type Factory struct {
	claude     ports.Installer
	cursor     ports.Installer
	vscode     ports.Installer
	claudeCode ports.Installer
	docker     ports.Installer
}

func NewFactory() factories.InstallerFactory {
	return &Factory{
		claude:     claude.New(),
		cursor:     cursor.New(),
		vscode:     vscode.New(),
		claudeCode: claudecode.New(),
		docker:     docker.New(),
	}
}

func (f *Factory) GetInstaller(ctx context.Context, clientType installer.ClientType) (ports.Installer, error) {
	switch clientType {
	case installer.ClientTypeClaude:
		return f.claude, nil
	case installer.ClientTypeCursor:
		return f.cursor, nil
	case installer.ClientTypeVSCode:
		return f.vscode, nil
	case installer.ClientTypeClaudeCode:
		return f.claudeCode, nil
	case installer.ClientTypeDocker:
		return f.docker, nil
	default:
		return nil, errors.ErrClientNotSupported
	}
}