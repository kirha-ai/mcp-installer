package installer

import (
	"context"
	"errors"
	"testing"

	"go.kirha.ai/mcp-installer/internal/core/domain/installer"
	"go.kirha.ai/mcp-installer/internal/core/ports"
)

type MockInstaller struct {
	configPath       string
	config           interface{}
	isRunning        bool
	shouldFailLoad   bool
	shouldFailSave   bool
	shouldFailAdd    bool
	shouldFailBackup bool
	backupPath       string
	hasServer        bool
}

func (m *MockInstaller) GetConfigPath() (string, error) {
	if m.configPath == "" {
		return "/mock/config/path", nil
	}
	return m.configPath, nil
}

func (m *MockInstaller) LoadConfig(ctx context.Context) (interface{}, error) {
	if m.shouldFailLoad {
		return nil, errors.New("mock load error")
	}
	if m.config == nil {
		return map[string]interface{}{}, nil
	}
	return m.config, nil
}

func (m *MockInstaller) AddMcpServer(ctx context.Context, config interface{}, server *installer.McpServer) (interface{}, error) {
	if m.shouldFailAdd {
		return nil, errors.New("mock add error")
	}
	return config, nil
}

func (m *MockInstaller) RemoveMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (interface{}, error) {
	return config, nil
}

func (m *MockInstaller) SaveConfig(ctx context.Context, config interface{}) error {
	if m.shouldFailSave {
		return errors.New("mock save error")
	}
	return nil
}

func (m *MockInstaller) BackupConfig(ctx context.Context) (string, error) {
	if m.shouldFailBackup {
		return "", errors.New("mock backup error")
	}
	return m.backupPath, nil
}

func (m *MockInstaller) RestoreConfig(ctx context.Context, backupPath string) error {
	return nil
}

func (m *MockInstaller) ValidateConfig(ctx context.Context, config interface{}) error {
	return nil
}

func (m *MockInstaller) IsClientRunning(ctx context.Context) (bool, error) {
	return m.isRunning, nil
}

func (m *MockInstaller) HasMcpServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (bool, error) {
	return m.hasServer, nil
}

func (m *MockInstaller) GetMcpServerConfig(ctx context.Context, config interface{}, vertical installer.VerticalType) (*installer.McpServer, error) {
	if !m.hasServer {
		return nil, errors.New("server not found")
	}
	return &installer.McpServer{
		Name:        "kirha-" + vertical.String(),
		Command:     "npx",
		Args:        []string{"-y", "@kirha/mcp-gateway"},
		Environment: map[string]string{"KIRHA_API_KEY": "test-key", "KIRHA_VERTICAL": vertical.String()},
	}, nil
}

func (m *MockInstaller) FormatConfig(ctx context.Context, config interface{}, onlyKirha bool) (string, error) {
	if !m.hasServer {
		return "No MCP servers configured", nil
	}
	return "Server: kirha-crypto\n  Command: npx\n  Args: [-y @kirha/mcp-gateway]\n  Environment:\n    KIRHA_API_KEY: test****", nil
}

func (m *MockInstaller) FormatSpecificServer(ctx context.Context, config interface{}, vertical installer.VerticalType) (string, error) {
	if !m.hasServer {
		return "No MCP servers configured", nil
	}
	return "Server: kirha-" + vertical.String() + "\n  Command: npx\n  Args: [-y @kirha/mcp-gateway]\n  Environment:\n    KIRHA_API_KEY: test****", nil
}

type MockFactory struct {
	installer ports.Installer
}

func (f *MockFactory) GetInstaller(ctx context.Context, clientType installer.ClientType) (ports.Installer, error) {
	return f.installer, nil
}

func TestApplication_Execute_Install_Success(t *testing.T) {
	mockInstaller := &MockInstaller{
		configPath: "/test/config.json",
		backupPath: "/test/config.json.backup",
		hasServer:  false, // No existing server for install
	}

	mockFactory := &MockFactory{installer: mockInstaller}
	app := New(mockFactory)

	config := &installer.Config{
		Client:    installer.ClientTypeClaude,
		Vertical:  installer.VerticalTypeCrypto,
		ApiKey:    "valid-api-key-123",
		Operation: installer.OperationInstall,
		DryRun:    false,
	}

	ctx := context.Background()
	result, err := app.Execute(ctx, config)

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if !result.Success {
		t.Errorf("Execute().Success = %v, want true", result.Success)
	}

	if result.ConfigPath != "/test/config.json" {
		t.Errorf("Execute().ConfigPath = %v, want /test/config.json", result.ConfigPath)
	}
}

func TestApplication_Execute_Show_Success(t *testing.T) {
	mockInstaller := &MockInstaller{
		configPath: "/test/config.json",
		hasServer:  true, // Has server for show
	}

	mockFactory := &MockFactory{installer: mockInstaller}
	app := New(mockFactory)

	config := &installer.Config{
		Client:    installer.ClientTypeClaude,
		Operation: installer.OperationShow,
		DryRun:    false,
	}

	ctx := context.Background()
	result, err := app.Execute(ctx, config)

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if !result.Success {
		t.Errorf("Execute().Success = %v, want true", result.Success)
	}

	expectedMessage := "MCP configuration for claude:\n\nServer: kirha-crypto\n  Command: npx\n  Args: [-y @kirha/mcp-gateway]\n  Environment:\n    KIRHA_API_KEY: test****"
	if result.Message != expectedMessage {
		t.Errorf("Execute().Message = %v, want %v", result.Message, expectedMessage)
	}
}

func TestApplication_Execute_Show_NoServer(t *testing.T) {
	mockInstaller := &MockInstaller{
		configPath: "/test/config.json",
		hasServer:  false, // No server for show
	}

	mockFactory := &MockFactory{installer: mockInstaller}
	app := New(mockFactory)

	config := &installer.Config{
		Client:    installer.ClientTypeClaude,
		Operation: installer.OperationShow,
		DryRun:    false,
	}

	ctx := context.Background()
	result, err := app.Execute(ctx, config)

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if !result.Success {
		t.Errorf("Execute().Success = %v, want true", result.Success)
	}

	expectedMessage := "No MCP servers configured for claude"
	if result.Message != expectedMessage {
		t.Errorf("Execute().Message = %v, want %v", result.Message, expectedMessage)
	}
}

func TestApplication_validateApiKey(t *testing.T) {
	app := &Application{}

	testCases := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{"valid key", "valid-api-key-123456", false},
		{"empty key", "", true},
		{"too short", "short", true},
		{"with spaces", "api key with spaces", true},
		{"whitespace only", "   ", true},
		{"leading/trailing spaces", "  valid-key  ", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := app.validateApiKey(tc.apiKey)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateApiKey() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
