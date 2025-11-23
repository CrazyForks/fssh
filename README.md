# fssh

![Touch ID Fingerprint](images/finger.png)

A solution for macOS that eliminates the need to manually unlock private keys with passphrases every time you use public key authentication, and helps you quickly view and connect to SSH hosts configured in `~/.ssh/config` without having to open the config file each time.

When SSH connects to a remote host, it prompts for fingerprint authentication, which decrypts and imports the private key upon successful verification. Additionally, fssh provides an interactive shell where you can quickly view and connect to hosts defined in `~/.ssh/config`.

## Use Cases
- Plaintext SSH keys stored locally on disk pose security risks
- Encrypted keys require entering passphrases every time you connect, which is inconvenient
- When multiple hosts are configured in `~/.ssh/config`, their aliases are easy to forget over time, requiring you to check the config file before each connection

## Solutions
- Automatically prompts for Touch ID fingerprint verification when using SSH commands to connect to hosts on macOS, then uses the master key to decrypt SSH private keys for authentication
- Compatible with OpenSSH's ssh-agent
- Running `fssh` directly launches an interactive shell where you can use commands like `list` and `connect` to view host information from `~/.ssh/config`. In this shell, you can enter a host's ID, host name, or IP directly to connect via SSH

## Screenshots
SSH connection with fingerprint unlock:

![Touch ID authentication](images/finger.png)

Viewing hosts from `~/.ssh/config` in the interactive shell:

![Interactive shell](images/shell.png)

Connecting to a host from the interactive shell:

![SSH connection](images/login.png)

## Features
- Touch ID/Local authentication to read master key (stored in Keychain)
- AES-256-GCM + HKDF encryption for private keys (independent `salt`/`nonce` per file)
- Private key import/export (PKCS#8 PEM backup), listing and status checking
- ssh-agent:
  - Supports fingerprint verification per signature, or configure TTL for repeat access within a short time window
  - Compatible with RSA-SHA2 (rsa-sha2-256/512)
- Interactive Shell: parses `~/.ssh/config` hosts, Tab completion, default connection behavior
- Configuration generator: automatically generates local `~/.ssh/config` entries with `IdentityAgent`

## Installation & Configuration
1. Build: `go build ./cmd/fssh`; install to `/usr/local/bin/fssh`
2. Run `fssh init` to initialize the master key
3. Run `fssh import -alias <string> -file <path/to/private> --ask-passphrase` to import private keys
4. Start the SSH authentication agent: `fssh agent --unlock-ttl-seconds 600`
5. Modify the `~/.ssh/config` file so all SSH connections go through the fssh agent (you can also use `export SSH_AUTH_SOCK=~/.fssh/agent.sock` to set the variable for specific use cases)

Configure the OpenSSH agent by adding the following at the beginning of `~/.ssh/config`:
```
host *
  ServerAliveInterval 30
  AddKeysToAgent yes
  ControlPersist 60
  ControlMaster auto
  IdentityAgent  ~/.fssh/agent.sock
```

6. If needed, create and modify the configuration file: `~/.fssh/config.json`, example:
```
{
    "socket":"~/.fssh/agent.sock",
    "require_touch_id_per_sign":true,
    "unlock_ttl_seconds":600,
    "log_level":"info",
    "log_format":"plain"
}
```

- socket: SSH agent socket location
- require_touch_id_per_sign: Whether to require Touch ID verification on every SSH signature
  - true: Security mode enabled, requires Touch ID on each signature (or after TTL expires)
  - false: Convenience mode, decrypts all keys once during startup and keeps them in memory
- unlock_ttl_seconds: Cache time window after Touch ID unlock
- log_level: Controls log output level
  - debug: Shows all logs (including cache hit information)
  - info: Shows general information (default)
  - warn: Shows warnings and errors only
  - error: Shows only errors
- log_format: Controls log output format
  - plain: Human-readable plain format
  - json: Structured JSON format

## Auto-start on Login
- Start agent: `fssh agent --unlock-ttl-seconds 600`
- Auto-start: Copy `contrib/com.fssh.agent.plist` to `~/Library/LaunchAgents/` and run `launchctl load -w`; after modifying configuration, use `launchctl kickstart -k gui/$(id -u)/com.fssh.agent` to reload

## Interactive Shell
- Launch: `fssh` or `fssh shell`
- Commands:
  - `list` displays `id\thost(ip)`
  - `search <term>` filters by id/host/ip
  - `connect <id|host|ip>` initiates connection; non-command input defaults to connection
  - Tab completion covers commands and id/host/ip

## Troubleshooting
- `"incorrect signature type / no mutual signature supported"`
  - Confirm agent is running and set `SSH_AUTH_SOCK=~/.fssh/agent.sock`
  - Local entries should include `IdentityAgent ~/.fssh/agent.sock`
  - Server accepts RSA-SHA2 (use `sshd-align` or manually edit)
- Input invisible after connection: use `ssh -tt` and suspend line editing during remote session
- Logging: configure `log_out/log_err`, `log_level`, `log_format`; restart agent after changes

## Security Notes
- Security mode: per-signature unlock (or TTL cache) avoids long-term decrypted private keys in memory
- Convenience mode: decrypts and keeps in memory on startup; only use this when prompted too frequently
- Avoid plaintext leakage: don't store plaintext keys/passwords in repositories or logs

## Credits
- This project is assisted by TRAE AI software
