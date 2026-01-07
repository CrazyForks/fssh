package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const plistLabel = "com.fssh.agent"

// setupLaunchAgent configures macOS LaunchAgent for auto-start
func setupLaunchAgent() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	launchAgentsDir := filepath.Join(home, "Library", "LaunchAgents")
	plistPath := filepath.Join(launchAgentsDir, plistLabel+".plist")

	// Check if already exists
	if _, err := os.Stat(plistPath); err == nil {
		fmt.Printf("⚠️  LaunchAgent already exists at %s\n", plistPath)

		// Try to unload first
		_ = exec.Command("launchctl", "unload", plistPath).Run()
	}

	// Ensure directory exists
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Generate plist content
	plistContent := generatePlistContent()

	// Write plist file
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	fmt.Printf("✓ Created plist: %s\n", plistPath)

	// Load LaunchAgent
	cmd := exec.Command("launchctl", "load", plistPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load LaunchAgent: %w", err)
	}

	fmt.Println("✓ LaunchAgent loaded successfully")

	return nil
}

// generatePlistContent generates the plist file content
func generatePlistContent() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>com.fssh.agent</string>
    <key>ProgramArguments</key>
    <array>
      <string>/usr/local/bin/fssh</string>
      <string>agent</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/fssh-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/fssh-agent.log</string>
  </dict>
</plist>
`
}
