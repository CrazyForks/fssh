# fssh — Touch ID–Protected SSH Agent and CLI

## Use Cases
- Plaintext SSH keys on disk are risky and easy to exfiltrate
- Encrypted keys require entering passphrases for every SSH login, which is inconvenient
- With many hosts in `~/.ssh/config`, aliases are easy to forget; you often need to open the file before every connection

## Solution (fssh — finger‑ssh)
- On macOS, unlock a Touch ID–protected master key to decrypt local encrypted SSH keys for authentication
- Provide an OpenSSH‑compatible ssh‑agent
- An interactive shell that parses hosts from `~/.ssh/config`, supports search and direct connect

## Overview
- Securely store and use SSH private keys on macOS with Touch ID (or equivalent local authentication)
- Provide an SSH agent that can operate in two modes:
  - Secure per‑sign unlock: prompts for Touch ID on each signature (or within a configurable TTL window)
  - Convenience preload: decrypts keys once and keeps them in memory for subsequent signatures
- Interactive shell to discover hosts from `~/.ssh/config` and connect fast with tab completion
- macOS login auto‑start via `launchd` and a generic LaunchAgent plist

## Key Features
- Touch ID‑protected master key, stored in macOS Keychain
- Encrypted key store at `~/.fssh/keys/<alias>.enc` using PKCS#8 + AES‑GCM
- RSA‑SHA2 signatures (`rsa‑sha2‑256/512`) supported by the agent
- Multiple keys import with unique `alias`
- Configurable agent socket and logging via `~/.fssh/config.json`
- Optional Touch ID TTL (`unlock_ttl_seconds`) to avoid repeated prompts in secure mode
- `config-gen` to generate local `~/.ssh/config` entries with `IdentityAgent`
- `sshd-align` to align server‑side `sshd_config` for RSA‑SHA2 algorithms

## Install (macOS)
- Build: `go build ./cmd/fssh`
- Place binary: `mv fssh /usr/local/bin/`
- Initialize: `fssh init`
- Import keys: `fssh import --alias work --file ~/.ssh/id_ed25519 --ask-passphrase`
- Config ssh agent,File: `~/.ssh/config`
- Example:
```
host *
  ServerAliveInterval 30
  AddKeysToAgent yes
  ControlPersist 60
  ControlMaster auto
  ControlPath ~/.ssh/shareconn/master-%r@%h:%p
  Ciphers +aes128-cbc,3des-cbc,aes192-cbc,aes256-cbc
  HostKeyAlgorithms +ssh-rsa
  KexAlgorithms +diffie-hellman-group1-sha1
  IdentityAgent  /Users/leo/.fssh/agent.sock
```

## Configuration
- File: `~/.fssh/config.json`
- Example:
```
{
  "socket": "~/.fssh/agent.sock",
  "require_touch_id_per_sign": true,
  "unlock_ttl_seconds": 600,
  "log_out": "/var/tmp/fssh-agent.out.log",
  "log_err": "/var/tmp/fssh-agent.err.log",
  "log_level": "info",
  "log_format": "plain",
  "log_time_format": "2006-01-02T15:04:05Z07:00"
}
```
- Precedence: CLI flags > config file > defaults

## Start Agent
- Foreground: `fssh agent`
- With TTL: `fssh agent --unlock-ttl-seconds 600`
- Use in shell: `export SSH_AUTH_SOCK=~/.fssh/agent.sock`

## Auto‑Start on Login (LaunchAgents)
- Copy plist: `cp contrib/com.fssh.agent.plist ~/Library/LaunchAgents/com.fssh.agent.plist`
- Load: `launchctl load -w ~/Library/LaunchAgents/com.fssh.agent.plist`
- Reload after config changes:
  - Preferred: `launchctl kickstart -k gui/$(id -u)/com.fssh.agent`
  - Legacy: `launchctl unload -w ~/Library/LaunchAgents/com.fssh.agent.plist && launchctl load -w ~/Library/LaunchAgents/com.fssh.agent.plist`

## Interactive Shell
- Start: `fssh` or `fssh shell`
- Commands:
  - `list` — show `id\thost(ip)`
  - `search <term>` — filter by id/host/ip
  - `connect <id|host|ip>` — connect via OpenSSH
  - Tab completion for commands and host/id/ip
  - Non‑command input defaults to `connect`

## Config Generator
- Print block: `fssh config-gen --host backuphost --user root`
- Write to file: `fssh config-gen --host backuphost --user root --write`
- Overwrite existing: `fssh config-gen --host backuphost --overwrite --write`
- Global algorithms (optional): `fssh config-gen --global-algos --write` adds RSA‑SHA2 once to `Host *`
- Generated host block contains `IdentityAgent` and optional `User/Port`; per‑host algorithm lines are not added by default

## Server Alignment (Optional)
- Align RSA‑SHA2 on server: `fssh sshd-align --host backuphost --sudo`
- Changes on remote `/etc/ssh/sshd_config`:
  - `PubkeyAuthentication yes`
  - `PubkeyAcceptedAlgorithms +rsa-sha2-512,rsa-sha2-256`
  - `PubkeyAcceptedKeyTypes +rsa-sha2-512,rsa-sha2-256` (compat)

## Troubleshooting
- “incorrect signature type / no mutual signature supported”
  - Ensure agent is running and environment: `export SSH_AUTH_SOCK=~/.fssh/agent.sock`
  - Client config for host uses agent: `IdentityAgent ~/.fssh/agent.sock`
  - Server accepts RSA‑SHA2 (use `sshd-align` or edit `sshd_config`)
- Input not visible after connect
  - Agent shell uses `ssh -tt` and suspends line editor during remote session
- Logging
  - Configure `log_out/log_err`, `log_level`, `log_format`; restart agent after changes

## Security Notes
- Secure mode: per‑sign unlock (or TTL cache) reduces risk by avoiding long‑lived decrypted keys
- Convenience mode: preload all keys into memory; prefer only when prompts are impractical
- Never store plaintext secrets in the repo or logs; use Keychain and config paths

## Comparison to Secretive
- Storage model
  - Secretive: keys in Secure Enclave, non‑exportable by design
  - fssh: PKCS#8 encrypted files in `~/.fssh/keys` with Touch ID‑protected master key in Keychain; optional password‑protected PEM export for recovery
- Access control
  - Secretive: Touch ID/Apple Watch gate before key access; access notifications
  - fssh: LocalAuthentication on each signature or within a configurable TTL window; no notifications yet
- Hardware support
  - Secretive: Smart Card/YubiKey supported for Macs without SE
  - fssh: no smart card support yet (roadmap)
- Agent and algorithms
  - Secretive: signs with SE‑backed keys (non‑exportable)
  - fssh: OpenSSH‑compatible agent with RSA‑SHA2 (`rsa‑sha2‑256/512`) extended signatures; includes `sshd-align` to align server algorithms
- Developer and ops tools
  - Secretive: native app experience and Homebrew install
  - fssh: CLI and interactive shell (host parsing, tab completion, default connect), `config-gen` to write `IdentityAgent` entries, generic `launchd` auto‑start, unified logging
- Platform
  - Secretive: macOS with Secure Enclave
  - fssh: macOS today; planned cross‑platform support

## Credits
- This project is assisted by TRAE AI software

## License
- Proprietary project (example). Adjust this section as appropriate for your distribution.
