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
- **Plan mode support**: Enable/disable tool plan mode for enhanced AI assistance
- **Flexible updates**: Update configurations without requiring API key changes

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
npx @kirha/mcp-installer install --client claude --vertical crypto --key your-api-key-here

# Install for Docker
npx @kirha/mcp-installer install --client docker --vertical crypto --key your-api-key-here

# Install for Cursor IDE with plan mode enabled
npx @kirha/mcp-installer install --client cursor --vertical crypto --key your-api-key-here --enable-plan-mode
```

### Update

```bash
# Update API key for Claude Desktop
npx @kirha/mcp-installer update --client claude --vertical crypto --key your-new-api-key

# Enable plan mode without changing API key
npx @kirha/mcp-installer update --client claude --vertical crypto --enable-plan-mode

# Disable plan mode without changing API key
npx @kirha/mcp-installer update --client claude --vertical crypto --disable-plan-mode

# Update API key and enable plan mode
npx @kirha/mcp-installer update --client docker --vertical crypto --key your-new-api-key --enable-plan-mode

# Update configuration preserving existing settings
npx @kirha/mcp-installer update --client cursor --vertical crypto
```

### Remove

```bash
# Remove from VS Code
npx @kirha/mcp-installer remove --client vscode --vertical crypto

# Remove from Cursor
npx @kirha/mcp-installer remove --client cursor --vertical crypto
```

### Show Configuration

```bash
# Show current configuration for Claude Desktop
npx @kirha/mcp-installer show --client claude --vertical crypto

# Show configuration for Docker
npx @kirha/mcp-installer show --client docker --vertical crypto

# Show configuration for VS Code with verbose output
npx @kirha/mcp-installer show --client vscode --vertical crypto --verbose
```

### Commands

- `install` - Install MCP server (fails if already exists)
- `update` - Update existing MCP server configuration (preserves existing settings when not specified)
- `remove` - Remove MCP server from configuration
- `show` - Display current MCP server configuration

### Options

#### Common Options
- `--client, -c` - Client to operate on (required)
- `--vertical` - Vertical to operate on (crypto, utils) (required)
- `--config-path` - Custom configuration file path (optional)
- `--dry-run` - Show what would be changed without making changes (install/update/remove only)
- `--verbose, -v` - Enable verbose logging

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
