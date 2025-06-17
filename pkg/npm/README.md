# @kirha/mcp-installer

MCP Installer is a CLI tool that simplifies the installation of Kirha MCP (Model Context Protocol) server across multiple development environments.

## Installation

```bash
npx @kirha/mcp-installer install --client <client> --key <api-key>
```

## Supported Clients

- **claude** - Claude Desktop application
- **cursor** - Cursor IDE
- **vscode** - Visual Studio Code
- **claude-code** - Claude Code CLI tool
- **docker** - Docker Compose setup

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

### Remove

```bash
# Remove from Cursor
npx @kirha/mcp-installer remove --client cursor

# Remove from VS Code
npx @kirha/mcp-installer remove --client vscode
```

### Commands

- `install` - Install MCP server (fails if already exists)
- `update` - Update existing MCP server configuration 
- `remove` - Remove MCP server from configuration

### Options

- `--client, -c` - Client to operate on (required)
- `--key, -k` - API key for Kirha MCP server (required for install/update)
- `--config-path` - Custom configuration file path (optional)
- `--dry-run` - Show what would be changed without making changes
- `--verbose, -v` - Enable verbose logging

### Examples

```bash
# Dry run to see what would be changed
npx @kirha/mcp-installer install --client vscode --key your-api-key-here --dry-run

# Custom config path
npx @kirha/mcp-installer install --client claude --key your-api-key-here --config-path /custom/path/config.json

# Verbose output
npx @kirha/mcp-installer update --client docker --key your-api-key-here --verbose
```

## Configuration Locations

### Claude Desktop
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%/Claude/claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

### Cursor
- **macOS**: `~/Library/Application Support/Cursor/User/settings.json`
- **Windows**: `%APPDATA%/Cursor/User/settings.json`
- **Linux**: `~/.config/Cursor/User/settings.json`

### VS Code
- **macOS**: `~/Library/Application Support/Code/User/settings.json`
- **Windows**: `%APPDATA%/Code/User/settings.json`
- **Linux**: `~/.config/Code/User/settings.json`

### Claude Code
- **All platforms**: `~/.claude-code/config.json`

### Docker
- **All platforms**: `./docker-compose.yml` (or `./docker-compose.mcp.yml` if main file exists)

## Troubleshooting

### Client is Running
If you get an error that the client is currently running, please close the application and try again.

### Permission Denied
Make sure you have write permissions to the configuration directory.

### Binary Not Found
If the binary is not found, this may indicate an incomplete installation. Try reinstalling:

```bash
npm uninstall -g @kirha/mcp-installer
npx @kirha/mcp-installer --client <client> --key <api-key>
```

## Support

For issues and feature requests, please visit: https://go.kirha.ai/mcp-installer/issues

## License

MIT