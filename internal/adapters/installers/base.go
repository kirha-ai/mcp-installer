package installers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.kirha.ai/mcp-installer/internal/core/domain/errors"
)

const (
	// Platform-specific directory names
	MacOSLibraryDir    = "Library"
	MacOSAppSupportDir = "Application Support"
	WindowsAppDataDir  = "AppData"
	WindowsRoamingDir  = "Roaming"
	LinuxConfigDir     = ".config"

	// Environment variables
	EnvAppData       = "APPDATA"
	EnvXDGConfigHome = "XDG_CONFIG_HOME"
)

type BaseInstaller struct {
}

func NewBaseInstaller() *BaseInstaller {
	return &BaseInstaller{}
}

func (b *BaseInstaller) ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.ErrConfigNotFound
		}
		if os.IsPermission(err) {
			return nil, errors.ErrPermissionDenied
		}
		slog.Error("failed to read file", slog.String("error", err.Error()), slog.String("path", path))
		return nil, errors.ErrConfigReadFailed
	}
	return data, nil
}

func (b *BaseInstaller) WriteFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("failed to create directory", slog.String("error", err.Error()), slog.String("dir", dir))
		return errors.ErrConfigWriteFailed
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		if os.IsPermission(err) {
			return errors.ErrPermissionDenied
		}
		slog.Error("failed to write file", slog.String("error", err.Error()), slog.String("path", path))
		return errors.ErrConfigWriteFailed
	}
	return nil
}

func (b *BaseInstaller) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (b *BaseInstaller) CreateBackup(path string) (string, error) {
	if !b.FileExists(path) {
		return "", errors.ErrConfigNotFound
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.backup_%s", path, timestamp)

	if b.FileExists(backupPath) {
		return "", errors.ErrBackupExists
	}

	data, err := b.ReadFile(path)
	if err != nil {
		return "", err
	}

	if err := b.WriteFile(backupPath, data); err != nil {
		slog.Error("failed to write backup file", slog.String("error", err.Error()), slog.String("backup_path", backupPath))
		return "", errors.ErrConfigBackupFailed
	}

	slog.Info("created configuration backup",
		slog.String("original", path),
		slog.String("backup", backupPath))

	return backupPath, nil
}

func (b *BaseInstaller) RestoreBackup(backupPath, targetPath string) error {
	if !b.FileExists(backupPath) {
		return errors.ErrConfigNotFound
	}

	data, err := b.ReadFile(backupPath)
	if err != nil {
		return err
	}

	if err := b.WriteFile(targetPath, data); err != nil {
		slog.Error("failed to restore file", slog.String("error", err.Error()), slog.String("target", targetPath))
		return errors.ErrConfigRestoreFailed
	}

	slog.Info("restored configuration from backup",
		slog.String("backup", backupPath),
		slog.String("target", targetPath))

	return nil
}

func (b *BaseInstaller) GetHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return home, nil
}

func (b *BaseInstaller) GetPlatformConfigPath(appName, fileName string) (string, error) {
	home, err := b.GetHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, MacOSLibraryDir, MacOSAppSupportDir, appName, fileName), nil
	case "windows":
		appData := os.Getenv(EnvAppData)
		if appData == "" {
			appData = filepath.Join(home, WindowsAppDataDir, WindowsRoamingDir)
		}
		return filepath.Join(appData, appName, fileName), nil
	case "linux":
		configHome := os.Getenv(EnvXDGConfigHome)
		if configHome == "" {
			configHome = filepath.Join(home, LinuxConfigDir)
		}
		return filepath.Join(configHome, appName, fileName), nil
	default:
		return "", fmt.Errorf("%w: %s", errors.ErrPlatformNotSupported, runtime.GOOS)
	}
}

func (b *BaseInstaller) LoadJSONConfig(ctx context.Context, path string) (map[string]interface{}, error) {
	data, err := b.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		slog.ErrorContext(ctx, "failed to parse JSON config", slog.String("error", err.Error()))
		return nil, errors.ErrConfigInvalid
	}

	return config, nil
}

func (b *BaseInstaller) SaveJSONConfig(ctx context.Context, path string, config interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal JSON config", slog.String("error", err.Error()))
		return errors.ErrConfigInvalid
	}

	return b.WriteFile(path, data)
}

func (b *BaseInstaller) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
