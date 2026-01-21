package installer

import "time"

const (
	ServerName = "kirha"
	ServerURL  = "https://mcp.kirha.com"
)

type ClientType string

const (
	ClientTypeClaudecode ClientType = "claudecode"
	ClientTypeCursor     ClientType = "cursor"
	ClientTypeCodex      ClientType = "codex"
	ClientTypeOpencode   ClientType = "opencode"
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
	Force      bool
}

type McpServer struct {
	Name    string
	Type    string
	URL     string
	Headers map[string]string
}

func NewKirhaRemoteMcpServer(apiKey string) *McpServer {
	return &McpServer{
		Name: ServerName,
		Type: "http",
		URL:  ServerURL,
		Headers: map[string]string{
			"Authorization": "Bearer " + apiKey,
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
