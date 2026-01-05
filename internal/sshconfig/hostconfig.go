package sshconfig

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// HostConfig represents a complete SSH host configuration
type HostConfig struct {
	// Basic identity
	Name     string // Host alias
	Hostname string // Target hostname/IP

	// Authentication
	User          string   // SSH username
	Port          string   // SSH port (default 22)
	IdentityFile  []string // Direct identity files
	IdentityAgent string   // SSH agent socket path

	// Proxy configuration (mutually exclusive)
	ProxyCommand string // Full proxy command
	ProxyJump    string // SSH jump host

	// Additional fields
	ForwardAgent        string // yes/no
	ServerAliveInterval string // keep-alive interval

	// Internal metadata
	Comment    string // Inline comment from config
	LineNumber int    // Original line number (for debugging)
}

// ProxyType represents the type of proxy configuration
type ProxyType int

const (
	ProxyTypeNone ProxyType = iota
	ProxyTypeSocks5NC
	ProxyTypeSocks5NCAT
	ProxyTypeHTTP
	ProxyTypeJump
	ProxyTypeCustom
)

// ProxyConfig represents proxy settings
type ProxyConfig struct {
	Type     ProxyType
	Host     string
	Port     string
	Username string // For jump hosts
}

// ValidateHostConfig validates a host configuration
func ValidateHostConfig(cfg *HostConfig) error {
	var errs []string

	// Validate required fields
	if cfg.Name == "" {
		errs = append(errs, "host name is required")
	}
	if strings.ContainsAny(cfg.Name, " \t\n") {
		errs = append(errs, "host name cannot contain whitespace")
	}
	if cfg.Hostname == "" {
		errs = append(errs, "hostname is required")
	}

	// Validate port
	if cfg.Port != "" {
		port, err := strconv.Atoi(cfg.Port)
		if err != nil || port < 1 || port > 65535 {
			errs = append(errs, "port must be between 1 and 65535")
		}
	}

	// Validate proxy configuration (mutually exclusive)
	if cfg.ProxyCommand != "" && cfg.ProxyJump != "" {
		errs = append(errs, "cannot specify both ProxyCommand and ProxyJump")
	}

	// Validate ForwardAgent
	if cfg.ForwardAgent != "" {
		fa := strings.ToLower(cfg.ForwardAgent)
		if fa != "yes" && fa != "no" {
			errs = append(errs, "ForwardAgent must be 'yes' or 'no'")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// BuildProxyCommand constructs a ProxyCommand string for SOCKS5 proxies
func BuildProxyCommand(proxyType ProxyType, host, port string) string {
	switch proxyType {
	case ProxyTypeSocks5NC:
		return fmt.Sprintf("nc -X 5 -x %s:%s %%h %%p", host, port)
	case ProxyTypeSocks5NCAT:
		return fmt.Sprintf("ncat --proxy-type socks5 --proxy %s:%s %%h %%p", host, port)
	case ProxyTypeHTTP:
		return fmt.Sprintf("nc -X connect -x %s:%s %%h %%p", host, port)
	default:
		return ""
	}
}

// BuildProxyJump constructs a ProxyJump string
func BuildProxyJump(user, host string) string {
	if user != "" {
		return fmt.Sprintf("%s@%s", user, host)
	}
	return host
}

// ParseProxyCommand attempts to parse an existing ProxyCommand
func ParseProxyCommand(cmd string) (*ProxyConfig, error) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil, errors.New("empty proxy command")
	}

	cfg := &ProxyConfig{Type: ProxyTypeCustom}

	// Try to detect SOCKS5 with nc
	if strings.Contains(cmd, "nc -X 5 -x") {
		cfg.Type = ProxyTypeSocks5NC
		// Extract host:port pattern
		parts := strings.Fields(cmd)
		for i, part := range parts {
			if part == "-x" && i+1 < len(parts) {
				hostPort := parts[i+1]
				if idx := strings.Index(hostPort, ":"); idx > 0 {
					cfg.Host = hostPort[:idx]
					cfg.Port = hostPort[idx+1:]
				}
				break
			}
		}
	} else if strings.Contains(cmd, "ncat --proxy-type socks5") {
		cfg.Type = ProxyTypeSocks5NCAT
		// Extract from --proxy
		parts := strings.Fields(cmd)
		for i, part := range parts {
			if part == "--proxy" && i+1 < len(parts) {
				hostPort := parts[i+1]
				if idx := strings.Index(hostPort, ":"); idx > 0 {
					cfg.Host = hostPort[:idx]
					cfg.Port = hostPort[idx+1:]
				}
				break
			}
		}
	}

	return cfg, nil
}

// ParseProxyJump parses an existing ProxyJump configuration
func ParseProxyJump(jump string) (*ProxyConfig, error) {
	jump = strings.TrimSpace(jump)
	if jump == "" {
		return nil, errors.New("empty proxy jump")
	}

	cfg := &ProxyConfig{Type: ProxyTypeJump}

	// Format: [user@]host[:port]
	// Handle user@host
	if idx := strings.Index(jump, "@"); idx > 0 {
		cfg.Username = jump[:idx]
		jump = jump[idx+1:]
	}

	// Handle host:port
	if idx := strings.Index(jump, ":"); idx > 0 {
		cfg.Host = jump[:idx]
		cfg.Port = jump[idx+1:]
	} else {
		cfg.Host = jump
	}

	return cfg, nil
}

// ProxyTemplates provides common proxy command templates
var ProxyTemplates = map[string]string{
	"socks5-nc":   "nc -X 5 -x %s:%s %%h %%p",
	"socks5-ncat": "ncat --proxy-type socks5 --proxy %s:%s %%h %%p",
	"http-nc":     "nc -X connect -x %s:%s %%h %%p",
}
