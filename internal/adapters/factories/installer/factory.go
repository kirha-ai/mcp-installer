package installerfactory

import (
	"context"

	"go.kirha.ai/mcp-installer/internal/adapters/installers/claude"
	claudecode "go.kirha.ai/mcp-installer/internal/adapters/installers/claude-code"
	"go.kirha.ai/mcp-installer/internal/adapters/installers/cursor"
	"go.kirha.ai/mcp-installer/internal/adapters/installers/docker"
	"go.kirha.ai/mcp-installer/internal/adapters/installers/vscode"
	"go.kirha.ai/mcp-installer/internal/core/domain/errors"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
	"go.kirha.ai/mcp-installer/internal/core/ports"
	"go.kirha.ai/mcp-installer/internal/core/ports/factories"
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
