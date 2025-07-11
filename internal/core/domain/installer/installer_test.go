package installer

import (
	"testing"
)

func TestClientType_String(t *testing.T) {
	tests := []struct {
		name     string
		client   ClientType
		expected string
	}{
		{"Claude", ClientTypeClaude, "claude"},
		{"Cursor", ClientTypeCursor, "cursor"},
		{"VSCode", ClientTypeVSCode, "vscode"},
		{"Claude Code", ClientTypeClaudeCode, "claude-code"},
		{"Docker", ClientTypeDocker, "docker"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.client) != tt.expected {
				t.Errorf("ClientType.String() = %v, want %v", string(tt.client), tt.expected)
			}
		})
	}
}

func TestNewKirhaMcpServer(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		vertical       VerticalType
		enablePlanMode bool
		expectedPlanMode string
	}{
		{
			name:           "Plan mode disabled",
			apiKey:         "test-api-key-123",
			vertical:       VerticalTypeCrypto,
			enablePlanMode: false,
			expectedPlanMode: "false",
		},
		{
			name:           "Plan mode enabled",
			apiKey:         "test-api-key-456",
			vertical:       VerticalTypeCrypto,
			enablePlanMode: true,
			expectedPlanMode: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewKirhaMcpServer(tt.apiKey, tt.vertical, tt.enablePlanMode)

			expectedName := "kirha-crypto"
			if server.Name != expectedName {
				t.Errorf("NewKirhaMcpServer().Name = %v, want %v", server.Name, expectedName)
			}

			if server.Command != "npx" {
				t.Errorf("NewKirhaMcpServer().Command = %v, want %v", server.Command, "npx")
			}

			expectedArgs := []string{"-y", "@kirha/mcp-gateway"}
			if len(server.Args) != len(expectedArgs) {
				t.Errorf("NewKirhaMcpServer().Args length = %v, want %v", len(server.Args), len(expectedArgs))
			}

			for i, arg := range server.Args {
				if arg != expectedArgs[i] {
					t.Errorf("NewKirhaMcpServer().Args[%d] = %v, want %v", i, arg, expectedArgs[i])
				}
			}

			if server.Environment["KIRHA_API_KEY"] != tt.apiKey {
				t.Errorf("NewKirhaMcpServer().Environment[KIRHA_API_KEY] = %v, want %v", server.Environment["KIRHA_API_KEY"], tt.apiKey)
			}

			expectedVerticalID := VerticalIDs[tt.vertical]
			if server.Environment["VERTICAL_ID"] != expectedVerticalID {
				t.Errorf("NewKirhaMcpServer().Environment[VERTICAL_ID] = %v, want %v", server.Environment["VERTICAL_ID"], expectedVerticalID)
			}

			if server.Environment["TOOL_PLAN_MODE_ENABLED"] != tt.expectedPlanMode {
				t.Errorf("NewKirhaMcpServer().Environment[TOOL_PLAN_MODE_ENABLED] = %v, want %v", server.Environment["TOOL_PLAN_MODE_ENABLED"], tt.expectedPlanMode)
			}
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	config := &Config{
		Client: ClientTypeClaude,
		ApiKey: "test-key",
	}

	if config.Client != ClientTypeClaude {
		t.Errorf("Config.Client = %v, want %v", config.Client, ClientTypeClaude)
	}

	if config.ApiKey != "test-key" {
		t.Errorf("Config.ApiKey = %v, want %v", config.ApiKey, "test-key")
	}
}
