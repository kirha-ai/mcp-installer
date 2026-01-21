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
		{"Claudecode", ClientTypeClaudecode, "claudecode"},
		{"Cursor", ClientTypeCursor, "cursor"},
		{"Codex", ClientTypeCodex, "codex"},
		{"Opencode", ClientTypeOpencode, "opencode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.client) != tt.expected {
				t.Errorf("ClientType.String() = %v, want %v", string(tt.client), tt.expected)
			}
		})
	}
}

func TestNewKirhaRemoteMcpServer(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
	}{
		{
			name:   "Valid API key",
			apiKey: "test-api-key-123",
		},
		{
			name:   "Another API key",
			apiKey: "test-api-key-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewKirhaRemoteMcpServer(tt.apiKey)

			expectedName := "kirha"
			if server.Name != expectedName {
				t.Errorf("NewKirhaRemoteMcpServer().Name = %v, want %v", server.Name, expectedName)
			}

			if server.Type != "http" {
				t.Errorf("NewKirhaRemoteMcpServer().Type = %v, want %v", server.Type, "http")
			}

			if server.URL != ServerURL {
				t.Errorf("NewKirhaRemoteMcpServer().URL = %v, want %v", server.URL, ServerURL)
			}

			expectedAuth := "Bearer " + tt.apiKey
			if server.Headers["Authorization"] != expectedAuth {
				t.Errorf("NewKirhaRemoteMcpServer().Headers[Authorization] = %v, want %v", server.Headers["Authorization"], expectedAuth)
			}
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	config := &Config{
		Client: ClientTypeClaudecode,
		ApiKey: "test-key",
	}

	if config.Client != ClientTypeClaudecode {
		t.Errorf("Config.Client = %v, want %v", config.Client, ClientTypeClaudecode)
	}

	if config.ApiKey != "test-key" {
		t.Errorf("Config.ApiKey = %v, want %v", config.ApiKey, "test-key")
	}
}
