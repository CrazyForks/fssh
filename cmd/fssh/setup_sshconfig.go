package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fssh/internal/otp"
)

// addToSSHConfig adds fssh agent configuration to SSH config
func addToSSHConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sshConfigPath := filepath.Join(home, ".ssh", "config")
	socketPath := filepath.Join(home, ".fssh", "agent.sock")

	// Check if config file exists
	var existingContent []byte
	if _, err := os.Stat(sshConfigPath); err == nil {
		existingContent, err = os.ReadFile(sshConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read SSH config: %w", err)
		}

		// Check if fssh agent is already configured
		if strings.Contains(string(existingContent), socketPath) {
			fmt.Println("✓ SSH config already contains fssh agent configuration")
			return nil
		}

		// Check if there's already a global IdentityAgent
		if strings.Contains(string(existingContent), "IdentityAgent") {
			fmt.Println("⚠️  SSH config already contains IdentityAgent configuration")
			if !otp.PromptConfirm("Do you want to add fssh agent configuration anyway?") {
				fmt.Println("Skipped SSH config update")
				displayManualInstructions(socketPath)
				return nil
			}
		}
	}

	// Ask user if they want to update SSH config
	fmt.Println("To use fssh agent automatically, we can add configuration to ~/.ssh/config")
	if !otp.PromptConfirm("Do you want to update SSH config?") {
		fmt.Println("Skipped SSH config update")
		displayManualInstructions(socketPath)
		return nil
	}

	// Create backup
	if len(existingContent) > 0 {
		backupPath := fmt.Sprintf("%s.bak.%d", sshConfigPath, time.Now().Unix())
		if err := os.WriteFile(backupPath, existingContent, 0600); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("✓ Created backup: %s\n", backupPath)
	}

	// Prepare new configuration
	fsshConfig := fmt.Sprintf(`# fssh agent configuration
Host *
    IdentityAgent %s

`, socketPath)

	// Prepend fssh configuration
	newContent := fsshConfig
	if len(existingContent) > 0 {
		newContent += string(existingContent)
	}

	// Ensure .ssh directory exists
	sshDir := filepath.Dir(sshConfigPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Write updated config
	if err := os.WriteFile(sshConfigPath, []byte(newContent), 0600); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}

	fmt.Printf("✓ Updated SSH config: %s\n", sshConfigPath)

	return nil
}

// displayManualInstructions shows manual configuration instructions
func displayManualInstructions(socketPath string) {
	fmt.Println()
	fmt.Println("To use fssh agent manually, add this to your ~/.ssh/config:")
	fmt.Println()
	fmt.Println("  Host *")
	fmt.Printf("      IdentityAgent %s\n", socketPath)
	fmt.Println()
	fmt.Println("Or set the environment variable:")
	fmt.Printf("  export SSH_AUTH_SOCK=%s\n", socketPath)
	fmt.Println()
}
