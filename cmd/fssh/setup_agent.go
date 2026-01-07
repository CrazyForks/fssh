package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// startAgent starts the fssh agent and verifies it's running
func startAgent() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	socketPath := filepath.Join(home, ".fssh", "agent.sock")

	// Check if agent is already running
	if _, err := os.Stat(socketPath); err == nil {
		// Try to connect
		if conn, err := net.Dial("unix", socketPath); err == nil {
			conn.Close()
			fmt.Println("✓ fssh agent is already running")
			return nil
		}

		// Socket exists but not responding, remove it
		_ = os.Remove(socketPath)
	}

	// Agent will be started by LaunchAgent automatically
	// We just need to wait for it to start
	fmt.Println("Waiting for agent to start...")

	// Wait up to 10 seconds for socket to appear and be ready
	maxAttempts := 20
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(500 * time.Millisecond)

		if _, err := os.Stat(socketPath); err == nil {
			// Socket exists, try to connect
			if conn, err := net.Dial("unix", socketPath); err == nil {
				conn.Close()
				fmt.Println("✓ fssh agent started successfully")
				return nil
			}
		}

		// Show progress
		if (i+1)%4 == 0 {
			fmt.Print(".")
		}
	}

	fmt.Println()
	return fmt.Errorf("agent did not start within 10 seconds")
}
