package installer

import "time"

const (
	ServerName = "kirha"
)

type VerticalType string

const (
	VerticalTypeCrypto VerticalType = "crypto"
)

var VerticalIDs = map[VerticalType]string{
	VerticalTypeCrypto: "cfdbb30b-e0e5-42d3-91c4-baca44234e36",
}

func (v VerticalType) Valid() bool {
	switch v {
	case VerticalTypeCrypto:
		return true
	default:
		return false
	}
}

func (v VerticalType) String() string {
	return string(v)
}

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
	Vertical   VerticalType
	ApiKey     string
	ConfigPath string
	Operation  OperationType
	DryRun     bool
	Verbose    bool
	OnlyKirha  bool
}

type McpServer struct {
	Name        string
	Command     string
	Args        []string
	Environment map[string]string
}

func NewKirhaMcpServer(apiKey string, vertical VerticalType) *McpServer {
	serverName := ServerName + "-" + vertical.String()
	return &McpServer{
		Name:    serverName,
		Command: "npx",
		Args:    []string{"-y", "@kirha/mcp-server", "stdio"},
		Environment: map[string]string{
			"KIRHA_API_KEY":  apiKey,
			"KIRHA_VERTICAL": VerticalIDs[vertical],
		},
	}
}

func GetServerName(vertical VerticalType) string {
	return ServerName + "-" + vertical.String()
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
