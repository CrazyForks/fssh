package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fssh/internal/sshconfig"
	"fssh/internal/store"
	"github.com/peterh/liner"
)

// ShellContext holds state for shell commands
type ShellContext struct {
	infos        []sshconfig.HostInfo
	hosts        []string
	byName       map[string]sshconfig.HostInfo
	byHostname   map[string]sshconfig.HostInfo
	ipToName     map[string]string
	idToName     map[string]string
	hostnames    []string
	ips          []string
	ids          []string
	liner        *liner.State
	importedKeys []string
}

// cmdAdd adds a new SSH host configuration
func cmdAdd(ctx *ShellContext) error {
	fmt.Println("\n=== Add New SSH Host ===")

	cfg := &sshconfig.HostConfig{}

	// 1. Host Alias (required)
	for {
		line, err := ctx.liner.Prompt("Host alias (e.g., myserver): ")
		if err != nil {
			return err
		}
		alias := strings.TrimSpace(line)
		if alias == "" {
			fmt.Println("Error: Host alias is required")
			continue
		}
		// Check if already exists
		if _, exists := ctx.byName[alias]; exists {
			fmt.Printf("Warning: Host '%s' already exists. Use 'edit %s' to modify.\n", alias, alias)
			return nil
		}
		cfg.Name = alias
		break
	}

	// 2. Hostname/IP (required)
	for {
		line, err := ctx.liner.Prompt("Hostname or IP: ")
		if err != nil {
			return err
		}
		hostname := strings.TrimSpace(line)
		if hostname == "" {
			fmt.Println("Error: Hostname is required")
			continue
		}
		cfg.Hostname = hostname
		break
	}

	// 3. Username (optional)
	line, _ := ctx.liner.Prompt("Username [press Enter to skip]: ")
	cfg.User = strings.TrimSpace(line)

	// 4. Port (optional)
	line, _ = ctx.liner.Prompt("Port [22]: ")
	port := strings.TrimSpace(line)
	if port != "" && port != "22" {
		cfg.Port = port
	}

	// 5. Private Key Selection
	fmt.Println("\nPrivate Key Configuration:")
	fmt.Println("  1) Use fssh agent (recommended)")
	fmt.Println("  2) Direct IdentityFile path")
	fmt.Println("  3) None (password authentication)")
	line, _ = ctx.liner.Prompt("Choice [1]: ")
	choice := strings.TrimSpace(line)
	if choice == "" {
		choice = "1"
	}

	switch choice {
	case "1":
		// List imported keys
		if len(ctx.importedKeys) == 0 {
			fmt.Println("Warning: No keys imported yet. Run 'fssh import' first.")
			fmt.Println("Using fssh agent socket anyway (you can import keys later)")
		} else {
			fmt.Println("\nImported keys:")
			for _, key := range ctx.importedKeys {
				fmt.Printf("  - %s\n", key)
			}
			fmt.Println("\nNote: All imported keys are available via the agent")
		}

		// Always set IdentityAgent for fssh usage
		home, _ := os.UserHomeDir()
		cfg.IdentityAgent = filepath.Join(home, ".fssh", "agent.sock")

	case "2":
		line, _ := ctx.liner.Prompt("IdentityFile path: ")
		idFile := strings.TrimSpace(line)
		if idFile != "" {
			cfg.IdentityFile = []string{idFile}
		}

	case "3":
		// No identity configuration
		fmt.Println("Using password authentication")
	}

	// 6. Proxy Configuration
	fmt.Println("\nProxy Configuration:")
	fmt.Println("  1) No proxy")
	fmt.Println("  2) SOCKS5 proxy")
	fmt.Println("  3) SSH jump host (ProxyJump)")
	fmt.Println("  4) Custom ProxyCommand")
	line, _ = ctx.liner.Prompt("Choice [1]: ")
	proxyChoice := strings.TrimSpace(line)
	if proxyChoice == "" {
		proxyChoice = "1"
	}

	switch proxyChoice {
	case "2":
		proxyHost, _ := ctx.liner.Prompt("SOCKS5 proxy host: ")
		proxyPort, _ := ctx.liner.Prompt("SOCKS5 proxy port [1080]: ")
		if proxyPort == "" {
			proxyPort = "1080"
		}

		// Detect which tool to use
		fmt.Println("  a) nc (netcat)")
		fmt.Println("  b) ncat (nmap ncat)")
		tool, _ := ctx.liner.Prompt("Tool [a]: ")

		proxyType := sshconfig.ProxyTypeSocks5NC
		if tool == "b" {
			proxyType = sshconfig.ProxyTypeSocks5NCAT
		}
		cfg.ProxyCommand = sshconfig.BuildProxyCommand(proxyType,
			strings.TrimSpace(proxyHost), proxyPort)

	case "3":
		jumpHost, _ := ctx.liner.Prompt("Jump host (user@host or just host): ")
		cfg.ProxyJump = strings.TrimSpace(jumpHost)

	case "4":
		proxyCmd, _ := ctx.liner.Prompt("ProxyCommand: ")
		cfg.ProxyCommand = strings.TrimSpace(proxyCmd)
	}

	// 7. Confirm and save
	fmt.Println("\n=== Summary ===")
	fmt.Printf("Host: %s\n", cfg.Name)
	fmt.Printf("Hostname: %s\n", cfg.Hostname)
	if cfg.User != "" {
		fmt.Printf("User: %s\n", cfg.User)
	}
	if cfg.Port != "" {
		fmt.Printf("Port: %s\n", cfg.Port)
	}
	if cfg.IdentityAgent != "" {
		fmt.Printf("IdentityAgent: %s\n", cfg.IdentityAgent)
	}
	for _, idFile := range cfg.IdentityFile {
		fmt.Printf("IdentityFile: %s\n", idFile)
	}
	if cfg.ProxyJump != "" {
		fmt.Printf("ProxyJump: %s\n", cfg.ProxyJump)
	}
	if cfg.ProxyCommand != "" {
		fmt.Printf("ProxyCommand: %s\n", cfg.ProxyCommand)
	}

	confirm, _ := ctx.liner.Prompt("\nSave this configuration? [Y/n]: ")
	if strings.ToLower(strings.TrimSpace(confirm)) == "n" {
		fmt.Println("Cancelled")
		return nil
	}

	// Write to SSH config
	if err := sshconfig.WriteHostConfig(cfg, false); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	// Reload context
	if err := reloadHosts(ctx); err != nil {
		fmt.Printf("Warning: Failed to reload hosts: %v\n", err)
	}

	fmt.Printf("\n✓ Host '%s' added to ~/.ssh/config\n", cfg.Name)
	fmt.Printf("✓ Backup created\n")
	fmt.Printf("\nYou can now connect with: ssh %s\n", cfg.Name)

	return nil
}

// cmdEdit edits an existing SSH host configuration
func cmdEdit(ctx *ShellContext, args string) error {
	hostName := strings.TrimSpace(args)
	if hostName == "" {
		fmt.Println("Usage: edit <host>")
		return nil
	}

	// Load existing configuration
	cfg, err := sshconfig.LoadHostConfig(hostName)
	if err != nil {
		return fmt.Errorf("host not found: %w", err)
	}

	fmt.Printf("\n=== Edit Host: %s ===\n", cfg.Name)
	fmt.Println("Press Enter to keep current value, or type new value")

	// Edit Hostname
	currentHostname := cfg.Hostname
	if currentHostname == "" {
		currentHostname = "(none)"
	}
	line, _ := ctx.liner.Prompt(fmt.Sprintf("Hostname [%s]: ", currentHostname))
	if line = strings.TrimSpace(line); line != "" {
		cfg.Hostname = line
	}

	// Edit User
	currentUser := cfg.User
	if currentUser == "" {
		currentUser = "(none)"
	}
	line, _ = ctx.liner.Prompt(fmt.Sprintf("User [%s]: ", currentUser))
	if line = strings.TrimSpace(line); line != "" {
		if line == "-" {
			cfg.User = "" // Clear field
		} else {
			cfg.User = line
		}
	}

	// Edit Port
	currentPort := cfg.Port
	if currentPort == "" {
		currentPort = "22"
	}
	line, _ = ctx.liner.Prompt(fmt.Sprintf("Port [%s]: ", currentPort))
	if line = strings.TrimSpace(line); line != "" {
		if line == "22" {
			cfg.Port = "" // Default port, no need to specify
		} else {
			cfg.Port = line
		}
	}

	// Edit Identity Configuration
	line, _ = ctx.liner.Prompt("\nChange identity configuration? [y/N]: ")
	if strings.ToLower(strings.TrimSpace(line)) == "y" {
		fmt.Println("  1) Use fssh agent")
		fmt.Println("  2) Direct IdentityFile")
		fmt.Println("  3) None")
		choice, _ := ctx.liner.Prompt("Choice: ")

		switch strings.TrimSpace(choice) {
		case "1":
			home, _ := os.UserHomeDir()
			cfg.IdentityAgent = filepath.Join(home, ".fssh", "agent.sock")
			cfg.IdentityFile = nil
		case "2":
			idFile, _ := ctx.liner.Prompt("IdentityFile path: ")
			cfg.IdentityFile = []string{strings.TrimSpace(idFile)}
			cfg.IdentityAgent = ""
		case "3":
			cfg.IdentityAgent = ""
			cfg.IdentityFile = nil
		}
	}

	// Edit Proxy Configuration
	line, _ = ctx.liner.Prompt("\nChange proxy configuration? [y/N]: ")
	if strings.ToLower(strings.TrimSpace(line)) == "y" {
		fmt.Println("  1) No proxy")
		fmt.Println("  2) SOCKS5 proxy")
		fmt.Println("  3) SSH jump host")
		fmt.Println("  4) Custom ProxyCommand")
		proxyChoice, _ := ctx.liner.Prompt("Choice: ")

		switch strings.TrimSpace(proxyChoice) {
		case "1":
			cfg.ProxyCommand = ""
			cfg.ProxyJump = ""

		case "2":
			proxyHost, _ := ctx.liner.Prompt("SOCKS5 proxy host: ")
			proxyPort, _ := ctx.liner.Prompt("SOCKS5 proxy port [1080]: ")
			if proxyPort == "" {
				proxyPort = "1080"
			}
			fmt.Println("  a) nc (netcat)")
			fmt.Println("  b) ncat (nmap ncat)")
			tool, _ := ctx.liner.Prompt("Tool [a]: ")

			proxyType := sshconfig.ProxyTypeSocks5NC
			if tool == "b" {
				proxyType = sshconfig.ProxyTypeSocks5NCAT
			}
			cfg.ProxyCommand = sshconfig.BuildProxyCommand(proxyType,
				strings.TrimSpace(proxyHost), proxyPort)
			cfg.ProxyJump = ""

		case "3":
			jumpHost, _ := ctx.liner.Prompt("Jump host (user@host or just host): ")
			cfg.ProxyJump = strings.TrimSpace(jumpHost)
			cfg.ProxyCommand = ""

		case "4":
			proxyCmd, _ := ctx.liner.Prompt("ProxyCommand: ")
			cfg.ProxyCommand = strings.TrimSpace(proxyCmd)
			cfg.ProxyJump = ""
		}
	}

	// Confirm and save
	confirm, _ := ctx.liner.Prompt("\nSave changes? [Y/n]: ")
	if strings.ToLower(strings.TrimSpace(confirm)) == "n" {
		fmt.Println("Cancelled")
		return nil
	}

	if err := sshconfig.WriteHostConfig(cfg, true); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	if err := reloadHosts(ctx); err != nil {
		fmt.Printf("Warning: Failed to reload hosts: %v\n", err)
	}

	fmt.Printf("\n✓ Host '%s' updated\n", cfg.Name)
	return nil
}

// cmdDelete deletes an SSH host configuration
func cmdDelete(ctx *ShellContext, args string) error {
	hostName := strings.TrimSpace(args)
	if hostName == "" {
		fmt.Println("Usage: delete <host>")
		return nil
	}

	// Check if host exists
	cfg, err := sshconfig.LoadHostConfig(hostName)
	if err != nil {
		return fmt.Errorf("host not found: %w", err)
	}

	// Show details and confirm
	fmt.Printf("\n=== Delete Host: %s ===\n", cfg.Name)
	fmt.Printf("Hostname: %s\n", cfg.Hostname)
	if cfg.User != "" {
		fmt.Printf("User: %s\n", cfg.User)
	}

	line, _ := ctx.liner.Prompt("\nType 'yes' to confirm deletion: ")
	if strings.TrimSpace(line) != "yes" {
		fmt.Println("Cancelled")
		return nil
	}

	if err := sshconfig.DeleteHostConfig(hostName); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	if err := reloadHosts(ctx); err != nil {
		fmt.Printf("Warning: Failed to reload hosts: %v\n", err)
	}

	fmt.Printf("\n✓ Host '%s' deleted from ~/.ssh/config\n", hostName)
	return nil
}

// cmdShow displays detailed host configuration
func cmdShow(ctx *ShellContext, args string) error {
	hostName := strings.TrimSpace(args)
	if hostName == "" {
		fmt.Println("Usage: show <host>")
		return nil
	}

	// Load configuration
	cfg, err := sshconfig.LoadHostConfig(hostName)
	if err != nil {
		return fmt.Errorf("host not found: %w", err)
	}

	// Display configuration
	fmt.Printf("\n=== Host: %s ===\n", cfg.Name)
	fmt.Printf("Hostname: %s\n", cfg.Hostname)

	if cfg.User != "" {
		fmt.Printf("User: %s\n", cfg.User)
	}
	if cfg.Port != "" {
		fmt.Printf("Port: %s\n", cfg.Port)
	} else {
		fmt.Printf("Port: 22 (default)\n")
	}

	if cfg.IdentityAgent != "" {
		fmt.Printf("IdentityAgent: %s\n", cfg.IdentityAgent)
	}
	for _, idFile := range cfg.IdentityFile {
		fmt.Printf("IdentityFile: %s\n", idFile)
	}

	if cfg.ProxyJump != "" {
		fmt.Printf("ProxyJump: %s\n", cfg.ProxyJump)
	}
	if cfg.ProxyCommand != "" {
		fmt.Printf("ProxyCommand: %s\n", cfg.ProxyCommand)
	}

	if cfg.ForwardAgent != "" {
		fmt.Printf("ForwardAgent: %s\n", cfg.ForwardAgent)
	}
	if cfg.ServerAliveInterval != "" {
		fmt.Printf("ServerAliveInterval: %s\n", cfg.ServerAliveInterval)
	}

	fmt.Println()
	return nil
}

// cmdInfo displays detailed host configuration by ID, alias, hostname, or IP
func cmdInfo(ctx *ShellContext, args string) error {
	query := strings.TrimSpace(args)
	if query == "" {
		fmt.Println("Usage: info <id|alias|hostname|ip>")
		return nil
	}

	// Resolve query to host alias
	hostName := resolveHostQuery(ctx, query)
	if hostName == "" {
		return fmt.Errorf("host not found: %s", query)
	}

	// Load configuration
	cfg, err := sshconfig.LoadHostConfig(hostName)
	if err != nil {
		return fmt.Errorf("failed to load host config: %w", err)
	}

	// Display configuration
	fmt.Printf("\n=== Host: %s ===\n", cfg.Name)
	fmt.Printf("Hostname: %s\n", cfg.Hostname)

	// Resolve IP if possible
	ip := resolveIPName(cfg.Hostname)
	if ip != "" {
		fmt.Printf("IP: %s\n", ip)
	}

	if cfg.User != "" {
		fmt.Printf("User: %s\n", cfg.User)
	}
	if cfg.Port != "" {
		fmt.Printf("Port: %s\n", cfg.Port)
	} else {
		fmt.Printf("Port: 22 (default)\n")
	}

	if cfg.IdentityAgent != "" {
		fmt.Printf("IdentityAgent: %s\n", cfg.IdentityAgent)
	}
	for _, idFile := range cfg.IdentityFile {
		fmt.Printf("IdentityFile: %s\n", idFile)
	}

	if cfg.ProxyJump != "" {
		fmt.Printf("ProxyJump: %s\n", cfg.ProxyJump)
	}
	if cfg.ProxyCommand != "" {
		fmt.Printf("ProxyCommand: %s\n", cfg.ProxyCommand)
	}

	if cfg.ForwardAgent != "" {
		fmt.Printf("ForwardAgent: %s\n", cfg.ForwardAgent)
	}
	if cfg.ServerAliveInterval != "" {
		fmt.Printf("ServerAliveInterval: %s\n", cfg.ServerAliveInterval)
	}

	fmt.Println()
	return nil
}

// resolveHostQuery resolves a query (id/alias/hostname/ip) to a host alias
func resolveHostQuery(ctx *ShellContext, query string) string {
	// Try ID first
	if name, ok := ctx.idToName[query]; ok {
		return name
	}

	// Try exact alias match
	if _, ok := ctx.byName[query]; ok {
		return query
	}

	// Try hostname match
	if hi, ok := ctx.byHostname[query]; ok {
		return hi.Name
	}

	// Try IP match
	if name, ok := ctx.ipToName[query]; ok {
		return name
	}

	return ""
}

// --- Helper functions ---

// listImportedKeys lists all imported private keys
func listImportedKeys() ([]string, error) {
	dir := store.KeysDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var keys []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".enc") {
			continue
		}
		// Remove .enc suffix to get alias
		alias := strings.TrimSuffix(e.Name(), ".enc")
		keys = append(keys, alias)
	}

	return keys, nil
}

// reloadHosts reloads host information into the context
func reloadHosts(ctx *ShellContext) error {
	// Reload host infos
	infos, err := sshconfig.LoadHostInfos()
	if err != nil {
		return err
	}

	// Rebuild all mappings
	ctx.infos = infos
	ctx.hosts = make([]string, 0, len(infos))
	ctx.byName = make(map[string]sshconfig.HostInfo)
	ctx.byHostname = make(map[string]sshconfig.HostInfo)
	ctx.ipToName = make(map[string]string)
	ctx.idToName = make(map[string]string)
	ctx.hostnames = []string{}
	ctx.ips = []string{}
	ctx.ids = []string{}

	for _, hi := range infos {
		ctx.hosts = append(ctx.hosts, hi.Name)
		ctx.byName[hi.Name] = hi
	}

	for _, hi := range infos {
		hn := displayHostname(hi)
		if hn != "" {
			ctx.byHostname[hn] = hi
			ctx.hostnames = append(ctx.hostnames, hn)
		}
	}

	for i, hi := range infos {
		id := fmt.Sprintf("%d", i+1)
		ctx.idToName[id] = hi.Name
		ctx.ids = append(ctx.ids, id)
	}

	for _, hi := range infos {
		hn := displayHostname(hi)
		ip := resolveIPName(hn)
		if ip != "" {
			ctx.ips = append(ctx.ips, ip)
			if _, ok := ctx.ipToName[ip]; !ok {
				ctx.ipToName[ip] = hi.Name
			}
		}
	}

	return nil
}
