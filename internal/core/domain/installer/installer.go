package installer

import "time"

const (
	ServerName = "kirha"
)

type ClientType string

const (
	ClientTypeClaude     ClientType = "claude"
	ClientTypeCursor     ClientType = "cursor"
	ClientTypeVSCode     ClientType = "vscode"
	ClientTypeClaudeCode ClientType = "claude-code"
	ClientTypeDocker     ClientType = "docker"
)

type OperationType string

const (
	OperationInstall OperationType = "install"
	OperationUpdate  OperationType = "update"
	OperationRemove  OperationType = "remove"
	OperationShow    OperationType = "show"
)

type Config struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time

	Client     ClientType
	ApiKey     string
	ConfigPath string
	Operation  OperationType
	DryRun     bool
	Verbose    bool
}

type McpServer struct {
	Name        string
	Command     string
	Args        []string
	Environment map[string]string
}

func NewKirhaMcpServer(apiKey string) *McpServer {
	return &McpServer{
		Name:    ServerName,
		Command: "npx",
		Args:    []string{"-y", "@kirha/mcp-server"},
		Environment: map[string]string{
			"KIRHA_API_KEY": apiKey,
		},
	}
}

type InstallResult struct {
	Success    bool
	ConfigPath string
	BackupPath string
	Message    string
}

type ShowResult struct {
	Success      bool
	ConfigPath   string
	HasServer    bool
	ServerConfig *McpServer
	FullConfig   string
	Message      string
}
