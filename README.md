# MCP Installer

[![Build Status](https://github.com/kirha-ai/mcp-installer/workflows/Build/badge.svg)](https://github.com/kirha-ai/mcp-installer/actions)
[![Go Reference](https://pkg.go.dev/badge/go.kirha.ai/mcp-installer.svg)](https://pkg.go.dev/go.kirha.ai/mcp-installer)
[![Go Report Card](https://goreportcard.com/badge/go.kirha.ai/mcp-installer)](https://goreportcard.com/report/go.kirha.ai/mcp-installer)
[![NPM Version](https://img.shields.io/npm/v/@kirha/mcp-installer)](https://www.npmjs.com/package/@kirha/mcp-installer)

MCP Installer is a CLI tool that simplifies the installation of Kirha MCP (Model Context Protocol) server across multiple development environments.

## Features

- **Multi-platform support**: Works on macOS, Linux, and Windows
- **Multiple client support**: Claude Code, Codex, OpenCode, Gemini CLI, and Droid (Factory AI)
- **Hexagonal Architecture**: Clean, maintainable, and testable codebase
- **Automatic backup**: Creates backups before modifying configurations
- **Dry-run mode**: Preview changes before applying them
- **Cross-platform builds**: Automated builds for multiple architectures

## Installation

### NPM (Recommended)

```bash
npx @kirha/mcp-installer install --client <client> --key <api-key>
```

### Go Run (Quick Start)

You can run the installer directly without downloading or installing:

```bash
go run go.kirha.ai/mcp-installer/cmd@latest install --client <client> --key <api-key>
```

### Direct Download

Download the latest binary from the [releases page](https://go.kirha.ai/mcp-installer/releases).

## Usage

### Install

```bash
# Install for Claude Code CLI
npx @kirha/mcp-installer install --client claudecode --key your-api-key-here

# Install for Codex
npx @kirha/mcp-installer install --client codex --key your-api-key-here

# Install for OpenCode
npx @kirha/mcp-installer install --client opencode --key your-api-key-here

# Install for Droid (Factory AI)
npx @kirha/mcp-installer install --client droid --key your-api-key-here

# Install for Gemini CLI (experimental)
npx @kirha/mcp-installer install --client gemini --key your-api-key-here

# Force install even if the client is running
npx @kirha/mcp-installer install --client claudecode --key your-api-key-here --force

# Using go run directly (without npm)
go run go.kirha.ai/mcp-installer/cmd@latest install --client claudecode --key your-api-key-here
```

### Update

```bash
# Update API key for Claude Code
npx @kirha/mcp-installer update --client claudecode --key your-new-api-key

# Update API key for Codex
npx @kirha/mcp-installer update --client codex --key your-new-api-key

# Force update even if the client is running
npx @kirha/mcp-installer update --client claudecode --key your-new-api-key --force
```

### Remove

```bash
# Remove from Claude Code
npx @kirha/mcp-installer remove --client claudecode

# Remove from Codex
npx @kirha/mcp-installer remove --client codex

# Force remove even if the client is running
npx @kirha/mcp-installer remove --client claudecode --force
```

### Show Configuration

```bash
# Show current configuration for Claude Code
npx @kirha/mcp-installer show --client claudecode

# Show configuration for Codex with verbose output
npx @kirha/mcp-installer show --client codex --verbose
```

### Commands

- `install` - Install MCP server (fails if already exists)
- `update` - Update existing MCP server configuration
- `remove` - Remove MCP server from configuration
- `show` - Display current MCP server configuration

### Options

#### Common Options
- `--client, -c` - Client to operate on (required)
- `--key, -k` - API key for the Kirha MCP server (required for install)
- `--config-path` - Custom configuration file path (optional)
- `--dry-run` - Show what would be changed without making changes (install/update/remove only)
- `--force, -f` - Force operation even if the client is running
- `--verbose` - Enable verbose logging

## Supported Clients

| Client | Status | Configuration Location |
|--------|--------|------------------------|
| **Claude Code** | Stable | `~/.claude.json` |
| **Codex** | Stable | `~/.codex/config.toml` |
| **OpenCode** | Stable | `~/.config/opencode/opencode.json` |
| **Droid** | Stable | `~/.factory/mcp.json` |
| **Gemini CLI** | Experimental* | `~/.gemini/settings.json` |

*Gemini CLI support is experimental due to server compatibility issues with Streamable HTTP transport.

## Development

### Prerequisites

- Go 1.22+
- Node.js 14+ (for NPM package)
- Wire (for dependency injection)

### Building from Source

```bash
# Clone the repository
git clone https://go.kirha.ai/mcp-installer.git
cd mcp-installer

# Install dependencies
go mod download

# Generate Wire code
go generate ./...

# Build for current platform
go build -o mcp-installer ./cmd

# Build for all platforms
make build-all
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Architecture

This project follows Hexagonal Architecture (Clean Architecture) principles:

```
├── cmd/                    # Application entry points
│   ├── cli/               # CLI commands (Cobra)
│   └── main.go           # Main entry point
├── di/                    # Dependency injection (Wire)
├── internal/              # Private application code
│   ├── adapters/         # External adapters (infrastructure)
│   │   ├── factories/    # Abstract factories
│   │   └── installers/   # Client-specific installers
│   ├── applications/     # Use cases/Application services
│   └── core/             # Business logic core
│       ├── domain/       # Domain entities and errors
│       └── ports/        # Interfaces/contracts
└── pkg/                  # Public/reusable packages
```

## License

MIT License - see [LICENSE](LICENSE) for details.
