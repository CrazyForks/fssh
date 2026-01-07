package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"fssh/internal/otp"
)

// ensureBinaryInstalled installs fssh to /usr/local/bin if not already installed
func ensureBinaryInstalled() error {
	targetPath := "/usr/local/bin/fssh"

	// Get current executable path
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Check if already installed at target location
	if _, err := os.Stat(targetPath); err == nil {
		// Target exists, check if it's the same file
		existingPath, err := filepath.EvalSymlinks(targetPath)
		if err == nil && existingPath == currentPath {
			fmt.Printf("✓ fssh is already installed at %s\n", targetPath)
			return nil
		}

		// Different file, ask if user wants to replace
		fmt.Printf("⚠️  fssh already exists at %s\n", targetPath)
		if !otp.PromptConfirm("Do you want to replace it?") {
			fmt.Println("Skipped binary installation")
			return nil
		}
	}

	// If we're already running from /usr/local/bin, no need to install
	if currentPath == targetPath {
		fmt.Printf("✓ Already running from %s\n", targetPath)
		return nil
	}

	// Inform user about sudo requirement
	fmt.Printf("Installing fssh to %s (requires sudo)...\n", targetPath)

	// Ensure /usr/local/bin exists
	if err := runSudoCommand("mkdir", "-p", "/usr/local/bin"); err != nil {
		return fmt.Errorf("failed to create /usr/local/bin: %w", err)
	}

	// Copy binary
	if err := runSudoCommand("cp", currentPath, targetPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Set permissions
	if err := runSudoCommand("chmod", "755", targetPath); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	fmt.Printf("✓ Successfully installed fssh to %s\n", targetPath)
	return nil
}

// runSudoCommand runs a command with sudo
func runSudoCommand(name string, args ...string) error {
	cmdArgs := append([]string{name}, args...)
	cmd := exec.Command("sudo", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
