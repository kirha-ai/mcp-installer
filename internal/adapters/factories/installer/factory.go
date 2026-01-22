package installerfactory

import (
	"context"

	"go.kirha.ai/mcp-installer/internal/adapters/installers/claudecode"
	"go.kirha.ai/mcp-installer/internal/adapters/installers/codex"
	"go.kirha.ai/mcp-installer/internal/adapters/installers/droid"
	"go.kirha.ai/mcp-installer/internal/adapters/installers/gemini"
	"go.kirha.ai/mcp-installer/internal/adapters/installers/opencode"
	"go.kirha.ai/mcp-installer/internal/core/domain/errors"
	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
	"go.kirha.ai/mcp-installer/internal/core/ports"
	"go.kirha.ai/mcp-installer/internal/core/ports/factories"
)

type Factory struct {
	claudecode ports.Installer
	codex      ports.Installer
	opencode   ports.Installer
	gemini     ports.Installer
	droid      ports.Installer
}

func NewFactory() factories.InstallerFactory {
	return &Factory{
		claudecode: claudecode.New(),
		codex:      codex.New(),
		opencode:   opencode.New(),
		gemini:     gemini.New(),
		droid:      droid.New(),
	}
}

func (f *Factory) GetInstaller(ctx context.Context, clientType installer.ClientType) (ports.Installer, error) {
	switch clientType {
	case installer.ClientTypeClaudecode:
		return f.claudecode, nil
	case installer.ClientTypeCodex:
		return f.codex, nil
	case installer.ClientTypeOpencode:
		return f.opencode, nil
	case installer.ClientTypeGemini:
		return f.gemini, nil
	case installer.ClientTypeDroid:
		return f.droid, nil
	default:
		return nil, errors.ErrClientNotSupported
	}
}
