# fssh - Secure and Simple SSH Key Management

![Touch ID Fingerprint](images/finger.png)

## What is fssh?

fssh is a macOS-only SSH key management tool that solves two common pain points:

1. **Entering private key passphrase every SSH login** → fssh lets you unlock with Touch ID or One-Time Password (OTP)
2. **Forgetting server aliases in `~/.ssh/config`** → fssh provides an interactive shell with Tab completion for quick connections

## How It Works

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  SSH Client │ ──▶ │  fssh Agent │ ──▶ │Remote Server│
└─────────────┘     └─────────────┘     └─────────────┘
                          │
                    Touch ID or OTP
                    unlocks encrypted
                    private keys
```

Your SSH private keys are stored encrypted. They can only be decrypted after Touch ID fingerprint or OTP verification.

## Screenshots

**Touch ID prompt during SSH login:**

![Touch ID Authentication](images/finger.png)

**Interactive shell for viewing and connecting to servers:**

![Interactive Shell](images/shell.png)

---

## Quick Start

### Option A: Interactive Setup Wizard (Recommended)

The easiest way to get started - one command does everything:

```bash
# Build the project
go build ./cmd/fssh

# Run interactive setup wizard
./fssh init
```

The wizard will guide you through:
1. **Choose authentication mode** - Touch ID or OTP
2. **Install binary** - Automatically copies fssh to /usr/local/bin
3. **Import SSH keys** - Scans and imports keys from ~/.ssh/
4. **Configure auto-start** - Sets up LaunchAgent for automatic startup
5. **Start agent** - Launches fssh agent immediately
6. **Configure SSH client** - Updates ~/.ssh/config automatically

After completion, you're ready to use SSH with Touch ID/OTP authentication!

### Option B: Manual Setup (Advanced)

For users who prefer step-by-step control:

#### Step 1: Install

```bash
# After downloading the source code, build it
go build ./cmd/fssh

# Install to system directory (requires admin privileges)
sudo cp fssh /usr/local/bin/
```

#### Step 2: Initialize

Choose an authentication mode based on your device:

**If your Mac has Touch ID (MacBook Pro/Air 2016+, iMac with Magic Keyboard, etc.):**

```bash
fssh init --mode touchid
```

**If your Mac doesn't have Touch ID (Mac Mini, older Macs, VMs, etc.):**

```bash
fssh init --mode otp
```

During OTP mode initialization:
1. You'll set a password (at least 12 characters)
2. A TOTP secret will be displayed - add it to an authenticator app (e.g., Google Authenticator, Authy)
3. 10 recovery codes will be shown - **save them securely**

#### Step 3: Import SSH Private Key

```bash
# Import your SSH private key (you'll be prompted for passphrase if the key has one)
fssh import --alias mykey --file ~/.ssh/id_rsa --ask-passphrase
```

Parameters:
- `--alias`: A name for this key for easy reference
- `--file`: Path to the private key file
- `--ask-passphrase`: Add this if the private key is passphrase-protected

#### Step 4: Start the Agent

```bash
fssh agent
```

Once started, the Agent runs in the background, listening on `~/.fssh/agent.sock`.

#### Step 5: Configure SSH to Use fssh Agent

Edit `~/.ssh/config` and add at the **very beginning**:

```
Host *
    IdentityAgent ~/.fssh/agent.sock
```

This routes all SSH connections through fssh Agent.

#### Step 6: Start Using

```bash
# Use SSH normally - Touch ID or OTP prompt will appear automatically
ssh user@yourserver.com
```

---

## Interactive Setup Wizard Details

The interactive wizard (`fssh init`) performs the following steps:

### Step-by-step Process

**Step 1: Welcome & Initialization Check**
- Displays welcome banner
- Checks if fssh is already initialized
- Prompts for confirmation if reinitializing

**Step 2: Choose Authentication Mode**
- Auto-detects Touch ID availability on your Mac
- Prompts you to choose between:
  - Touch ID (recommended for supported devices)
  - OTP (for devices without Touch ID or additional security)

**Step 3: Initialize Authentication**
- Executes the selected authentication mode setup
- For Touch ID: Generates and stores master key in macOS Keychain
- For OTP: Sets up password + TOTP with recovery codes

**Step 4: Binary Installation**
- Detects current executable location
- Copies fssh to `/usr/local/bin/` (requires sudo)
- Sets proper permissions (755)

**Step 5: Import SSH Keys**
- Scans `~/.ssh/` directory for private keys
- Detects standard key files: `id_rsa`, `id_ed25519`, `id_ecdsa`, etc.
- Identifies encrypted vs unencrypted keys
- Prompts for each key:
  - Suggested alias (e.g., `id_ed25519` → `ed25519`)
  - Passphrase if key is encrypted
- Imports keys with encryption protection

**Step 6: Configure LaunchAgent**
- Creates plist file at `~/Library/LaunchAgents/com.fssh.agent.plist`
- Loads LaunchAgent for auto-start on login
- Configures agent to keep running

**Step 7: Start Agent**
- Waits for agent to start (up to 10 seconds)
- Verifies socket connection at `~/.fssh/agent.sock`

**Step 8: Configure SSH Client**
- Prompts to update `~/.ssh/config`
- Creates backup before modification
- Prepends `IdentityAgent` configuration for automatic SSH integration

### Non-Interactive Mode

For automation, scripts, or CI/CD:

```bash
# Skip all interactive prompts
fssh init --non-interactive --mode touchid

# Or specify mode directly
fssh init --mode otp
```

---

## Auto-Start on Login

Tired of manually starting the Agent after each reboot? Set up auto-start:

```bash
# Copy the launch configuration file
cp contrib/com.fssh.agent.plist ~/Library/LaunchAgents/

# Load the service
launchctl load -w ~/Library/LaunchAgents/com.fssh.agent.plist
```

**Check service status:**

```bash
launchctl list | grep fssh
```

Normal output looks like: `-    0    com.fssh.agent` (0 means running)

**To restart the service:**

```bash
launchctl kickstart -k gui/$(id -u)/com.fssh.agent
```

**To stop the service:**

```bash
launchctl unload ~/Library/LaunchAgents/com.fssh.agent.plist
```

---

## Interactive Shell

Run `fssh` or `fssh shell` to enter interactive mode:

```bash
$ fssh
fssh> list                    # List all hosts from ~/.ssh/config
fssh> search prod             # Search for hosts containing "prod"
fssh> connect myserver        # Connect to myserver
fssh> myserver                # Or just type the hostname to connect
fssh> exit                    # Exit the shell
```

**Tab completion** is supported - type partial hostname and press Tab to autocomplete.

---

## Command Reference

### Initialization Commands

| Command | Description |
|---------|-------------|
| `fssh init` | Interactive setup wizard (recommended) |
| `fssh init --interactive` | Explicitly run interactive wizard |
| `fssh init --mode touchid` | Initialize with Touch ID (non-interactive) |
| `fssh init --mode otp` | Initialize with OTP (non-interactive) |
| `fssh init --non-interactive --mode touchid` | Non-interactive mode for scripts/CI |

### Key Management

| Command | Description |
|---------|-------------|
| `fssh import --alias name --file path --ask-passphrase` | Import a private key |
| `fssh list` | List imported keys |
| `fssh export --alias name --out path` | Export a key (backup) |
| `fssh remove --alias name` | Remove a key |

### Agent & Shell

| Command | Description |
|---------|-------------|
| `fssh agent` | Start the Agent |
| `fssh status` | Check status |
| `fssh shell` | Enter interactive shell |

---

## Configuration

Configuration file location: `~/.fssh/config.json`

```json
{
    "socket": "~/.fssh/agent.sock",
    "require_touch_id_per_sign": true,
    "unlock_ttl_seconds": 600,
    "log_level": "info",
    "log_format": "plain"
}
```

**Configuration options:**

| Option | Description | Default |
|--------|-------------|---------|
| `socket` | Agent socket path | `~/.fssh/agent.sock` |
| `require_touch_id_per_sign` | Require verification for each SSH signature (secure mode) | `true` |
| `unlock_ttl_seconds` | Cache duration after verification (seconds) - no re-verification needed within this period | `600` (10 min) |
| `log_level` | Log level: `debug`/`info`/`warn`/`error` | `info` |
| `log_format` | Log format: `plain` (readable) / `json` (structured) | `plain` |

**Secure Mode vs Convenience Mode:**

- `require_touch_id_per_sign: true` (Secure): Verification required for each SSH connection (or within TTL cache period)
- `require_touch_id_per_sign: false` (Convenience): All keys decrypted once at startup, no further verification needed

---

## Troubleshooting

### 1. Error "incorrect signature type" or "no mutual signature supported"

**Cause**: SSH client isn't using fssh Agent, or server doesn't support RSA-SHA2.

**Solution**:
1. Verify Agent is running: `launchctl list | grep fssh`
2. Verify `~/.ssh/config` has `IdentityAgent ~/.fssh/agent.sock`
3. Or set environment variable: `export SSH_AUTH_SOCK=~/.fssh/agent.sock`

### 2. No input display (cursor doesn't move)

**Cause**: Terminal control issue.

**Solution**: Use `ssh -tt` to force TTY allocation:

```bash
ssh -tt user@server
```

### 3. launchctl load error "Load failed: 5: Input/output error"

**Cause**: Service already loaded, or initialization incomplete.

**Solution**:

```bash
# Unload first
launchctl unload ~/Library/LaunchAgents/com.fssh.agent.plist

# Ensure initialization is complete
fssh init --mode touchid  # or otp

# Reload
launchctl load -w ~/Library/LaunchAgents/com.fssh.agent.plist
```

### 4. OTP verification code always wrong

**Cause**: Phone time out of sync, or TOTP was added incorrectly.

**Solution**:
1. Ensure phone time is correct (enable automatic time setting)
2. Delete the old entry in authenticator app and re-add
3. If recovery is impossible, use recovery codes

### 5. Forgot OTP password?

If you saved recovery codes, use them to reset. Without recovery codes, you must reinitialize (losing imported keys):

```bash
# Warning: This deletes all imported keys!
rm -rf ~/.fssh
fssh init --mode otp
# Then re-import your keys
```

---

## Security Notes

1. **Encrypted key storage**: Imported private keys are encrypted with AES-256-GCM - even if your computer is stolen, keys can't be decrypted without Touch ID/OTP
2. **Enable FileVault**: macOS full-disk encryption provides additional protection
3. **Protect recovery codes**: OTP recovery codes are like master keys - store them securely
4. **Regular backups**: Use `fssh export` to backup important keys

---

## Technical Details

- **Encryption**: AES-256-GCM + HKDF (independent salt/nonce per key file)
- **Key derivation**: PBKDF2 (100,000 iterations)
- **TOTP standard**: RFC 6238
- **Compatibility**: Fully compatible with OpenSSH ssh-agent protocol

---

## Credits

This project was developed with assistance from TRAE AI software.
