# MCP Installer

[![Build Status](https://go.kirha.ai/mcp-installer/workflows/Build%20and%20Test/badge.svg)](https://go.kirha.ai/mcp-installer/actions)
[![Go Report Card](https://goreportcard.com/badge/go.kirha.ai/mcp-installer)](https://goreportcard.com/report/go.kirha.ai/mcp-installer)
[![NPM Version](https://img.shields.io/npm/v/@kirha/mcp-installer)](https://www.npmjs.com/package/@kirha/mcp-installer)

MCP Installer is a CLI tool that simplifies the installation of Kirha MCP (Model Context Protocol) server across multiple development environments.

## Features

- **Multi-platform support**: Works on macOS, Linux, and Windows
- **Multiple client support**: Claude Desktop, Cursor, VS Code, Claude Code CLI, and Docker
- **Hexagonal Architecture**: Clean, maintainable, and testable codebase
- **Automatic backup**: Creates backups before modifying configurations
- **Dry-run mode**: Preview changes before applying them
- **Cross-platform builds**: Automated builds for multiple architectures

## Installation

### NPM (Recommended)

```bash
npx @kirha/mcp-installer install --client <client> --key <api-key>
```

### Direct Download

Download the latest binary from the [releases page](https://go.kirha.ai/mcp-installer/releases).

## Usage

### Install

```bash
# Install for Claude Desktop
npx @kirha/mcp-installer install --client claude --key your-api-key-here

# Install for Docker
npx @kirha/mcp-installer install --client docker --key your-api-key-here

# Install for Cursor IDE
npx @kirha/mcp-installer install --client cursor --key your-api-key-here
```

### Update

```bash
# Update configuration for Claude Desktop
npx @kirha/mcp-installer update --client claude --key your-new-api-key

# Update for Docker
npx @kirha/mcp-installer update --client docker --key your-new-api-key
```

7od0a2joxb3vgddq9g448msrhbkxbmla
### Remove

```bash
# Remove from VS Code
npx @kirha/mcp-installer remove --client vscode

# Remove from Cursor
npx @kirha/mcp-installer remove --client cursor
```

### Show Configuration

```bash
# Show current configuration for Claude Desktop
npx @kirha/mcp-installer show --client claude

# Show configuration for Docker
npx @kirha/mcp-installer show --client docker

# Show configuration for VS Code with verbose output
npx @kirha/mcp-installer show --client vscode --verbose
```

### Commands

- `install` - Install MCP server (fails if already exists)
- `update` - Update existing MCP server configuration
- `remove` - Remove MCP server from configuration
- `show` - Display current MCP server configuration

### Options

- `--client, -c` - Client to operate on (required)
- `--key, -k` - API key for Kirha MCP server (required for install/update)
- `--config-path` - Custom configuration file path (optional)
- `--dry-run` - Show what would be changed without making changes (install/update/remove only)
- `--verbose, -v` - Enable verbose logging

**Note**: The `show` command does not require an API key and will mask sensitive information for security.

## Supported Clients

| Client | Platform Support | Configuration Location |
|--------|------------------|------------------------|
| **Claude Desktop** | macOS, Windows, Linux | `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) |
| **Cursor** | macOS, Windows, Linux | `~/Library/Application Support/Cursor/User/settings.json` (macOS) |
| **VS Code** | macOS, Windows, Linux | `~/Library/Application Support/Code/User/settings.json` (macOS) |
| **Claude Code** | macOS, Windows, Linux | `~/.claude-code/config.json` |
| **Docker** | macOS, Windows, Linux | `./docker-compose.yml` (or `./docker-compose.mcp.yml`) |

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