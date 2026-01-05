package sshconfig

import (
	"fmt"
	"strings"
)

// LoadGlobalConfig loads the Host * configuration block
// Returns (config, found, error)
func LoadGlobalConfig() (*HostConfig, bool, error) {
	lines, err := readSSHConfigLines()
	if err != nil {
		return nil, false, err
	}

	start, end, found := findHostBlock(lines, "*")
	if !found {
		return nil, false, nil
	}

	cfg, err := parseHostBlock(lines, start, end)
	if err != nil {
		return nil, false, err
	}

	cfg.IsGlobal = true
	return cfg, true, nil
}

// WriteGlobalConfig writes or updates the Host * configuration
// Creates the block at the END of config file if it doesn't exist
func WriteGlobalConfig(cfg *HostConfig) error {
	// Mark as global
	cfg.Name = "*"
	cfg.IsGlobal = true

	// Validate configuration (allows missing Hostname)
	if err := ValidateHostConfig(cfg); err != nil {
		return fmt.Errorf("invalid global config: %w", err)
	}

	// Create backup
	backupPath, err := backupSSHConfig()
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Read current config
	lines, err := readSSHConfigLines()
	if err != nil {
		return err
	}

	// Find existing Host * block
	start, end, found := findHostBlock(lines, "*")

	// Render new Host * block
	newBlock := renderHostBlock(cfg)
	newBlockLines := strings.Split(strings.TrimRight(newBlock, "\n"), "\n")

	// Construct updated config
	var result []string
	if found {
		// Replace existing block IN PLACE
		result = append(result, lines[:start]...)
		result = append(result, newBlockLines...)
		result = append(result, lines[end:]...)
	} else {
		// Append to end of file (SSH best practice: Host * at the end)
		result = append(result, lines...)
		if len(result) > 0 && result[len(result)-1] != "" {
			result = append(result, "") // Blank line before Host *
		}
		result = append(result, newBlockLines...)
	}

	// Write updated config
	if err := writeSSHConfigLines(result); err != nil {
		// Restore backup on failure
		if backupPath != "" {
			configPath := sshConfigPath()
			_ = copyFile(backupPath, configPath)
		}
		return err
	}

	return nil
}

// SetGlobalOption sets a single global configuration option
// Creates Host * block if it doesn't exist
func SetGlobalOption(key, value string) error {
	// Normalize key (case-insensitive)
	key = normalizeConfigKey(key)

	// Validate option name
	if !IsValidGlobalOption(key) {
		return fmt.Errorf("unsupported global option: %s", key)
	}

	// Load existing config or create new one
	cfg, found, err := LoadGlobalConfig()
	if err != nil {
		return err
	}
	if !found {
		cfg = NewGlobalConfig()
	}

	// Set the option value
	if err := setConfigField(cfg, key, value); err != nil {
		return err
	}

	// Write back
	return WriteGlobalConfig(cfg)
}

// UnsetGlobalOption removes a global configuration option
func UnsetGlobalOption(key string) error {
	// Normalize key
	key = normalizeConfigKey(key)

	// Load existing config
	cfg, found, err := LoadGlobalConfig()
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("Host * block not found")
	}

	// Clear the option value
	if err := setConfigField(cfg, key, ""); err != nil {
		return err
	}

	// Write back
	return WriteGlobalConfig(cfg)
}

// normalizeConfigKey normalizes a config key to standard SSH format
// Converts various formats to the canonical SSH config key format
func normalizeConfigKey(key string) string {
	// Remove spaces and underscores
	key = strings.ReplaceAll(key, " ", "")
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")

	// Convert to lowercase for case-insensitive matching
	lower := strings.ToLower(key)

	// Try exact match first (case-insensitive)
	for standardKey := range GlobalConfigOptions {
		if strings.ToLower(standardKey) == lower {
			return standardKey
		}
	}

	// If no match, try to capitalize first letter
	if len(key) > 0 {
		return strings.ToUpper(key[:1]) + key[1:]
	}

	return key
}

// setConfigField sets a config field by name
func setConfigField(cfg *HostConfig, key, value string) error {
	value = strings.TrimSpace(value)

	switch key {
	case "ServerAliveInterval":
		cfg.ServerAliveInterval = value
	case "ServerAliveCountMax":
		cfg.ServerAliveCountMax = value
	case "ForwardAgent":
		cfg.ForwardAgent = value
	case "IdentityAgent":
		cfg.IdentityAgent = value
	case "AddKeysToAgent":
		cfg.AddKeysToAgent = value
	case "UseKeychain":
		cfg.UseKeychain = value
	case "PubkeyAcceptedAlgorithms":
		cfg.PubkeyAcceptedAlgorithms = value
	case "StrictHostKeyChecking":
		cfg.StrictHostKeyChecking = value
	case "UserKnownHostsFile":
		cfg.UserKnownHostsFile = value
	case "Compression":
		cfg.Compression = value
	case "TCPKeepAlive":
		cfg.TCPKeepAlive = value
	default:
		return fmt.Errorf("unsupported option: %s", key)
	}

	return ValidateHostConfig(cfg)
}

// getConfigField retrieves a config field value by name
func getConfigField(cfg *HostConfig, key string) string {
	switch key {
	case "ServerAliveInterval":
		return cfg.ServerAliveInterval
	case "ServerAliveCountMax":
		return cfg.ServerAliveCountMax
	case "ForwardAgent":
		return cfg.ForwardAgent
	case "IdentityAgent":
		return cfg.IdentityAgent
	case "AddKeysToAgent":
		return cfg.AddKeysToAgent
	case "UseKeychain":
		return cfg.UseKeychain
	case "PubkeyAcceptedAlgorithms":
		return cfg.PubkeyAcceptedAlgorithms
	case "StrictHostKeyChecking":
		return cfg.StrictHostKeyChecking
	case "UserKnownHostsFile":
		return cfg.UserKnownHostsFile
	case "Compression":
		return cfg.Compression
	case "TCPKeepAlive":
		return cfg.TCPKeepAlive
	default:
		return ""
	}
}
