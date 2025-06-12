//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	installerfactory "github.com/kirha-ai/mcp-installer/internal/adapters/factories/installer"
	"github.com/kirha-ai/mcp-installer/internal/applications/installer"
)

func ProvideInstallerApplication() (*installer.Application, error) {
	wire.Build(
		installerfactory.NewFactory,
		installer.New,
	)
	return nil, nil
}
