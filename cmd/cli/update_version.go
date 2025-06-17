package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdUpdateVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Kirha MCP Installer to the latest version",
		Long:  "Update Kirha MCP Installer to the latest version available on npm",
		RunE:  runUpdate,
	}

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Updating Kirha MCP Installer...")

	// Check if npm is available
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is not installed or not in PATH")
	}

	// Get the latest version first
	latest, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check latest version: %w", err)
	}

	if latest == Version {
		fmt.Printf("✅ Already on the latest version: %s\n", Version)
		return nil
	}

	fmt.Printf("📦 Updating from %s to %s...\n", Version, latest)

	// Run npm install -g @kirha/mcp-installer@latest
	updateCmd := exec.Command("npm", "install", "-g", "@kirha/mcp-installer@latest")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	fmt.Printf("✅ Successfully updated to version %s\n", latest)
	fmt.Println("Please restart your application to use the new version.")

	return nil
}
