#!/bin/bash
# fssh Environment Setup Script
# This script helps you set up SSH_AUTH_SOCK environment variable

set -e

# Detect socket path
SOCKET_PATH="$HOME/.fssh/agent.sock"
if [ -f "$HOME/.fssh/config.json" ]; then
    # Try to extract socket path from config
    CUSTOM_SOCKET=$(grep -o '"socket":\s*"[^"]*"' "$HOME/.fssh/config.json" | cut -d'"' -f4)
    if [ -n "$CUSTOM_SOCKET" ]; then
        # Expand ~ to $HOME
        SOCKET_PATH="${CUSTOM_SOCKET/#\~/$HOME}"
    fi
fi

echo "fssh Environment Setup"
echo "======================"
echo ""
echo "Socket path: $SOCKET_PATH"
echo ""

# Detect shell
SHELL_NAME=$(basename "$SHELL")
case "$SHELL_NAME" in
    bash)
        if [ -f "$HOME/.bash_profile" ]; then
            CONFIG_FILE="$HOME/.bash_profile"
        else
            CONFIG_FILE="$HOME/.bashrc"
        fi
        ;;
    zsh)
        CONFIG_FILE="$HOME/.zshrc"
        ;;
    fish)
        CONFIG_FILE="$HOME/.config/fish/config.fish"
        mkdir -p "$(dirname "$CONFIG_FILE")"
        ;;
    *)
        echo "Warning: Unknown shell '$SHELL_NAME'"
        CONFIG_FILE="$HOME/.profile"
        ;;
esac

echo "Detected shell: $SHELL_NAME"
echo "Config file: $CONFIG_FILE"
echo ""

# Check if already configured
if [ -f "$CONFIG_FILE" ] && grep -q "SSH_AUTH_SOCK.*fssh" "$CONFIG_FILE"; then
    echo "✓ SSH_AUTH_SOCK already configured in $CONFIG_FILE"
    echo ""
    echo "Current configuration:"
    grep "SSH_AUTH_SOCK.*fssh" "$CONFIG_FILE"
    echo ""
    read -p "Do you want to update it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "No changes made."
        exit 0
    fi
    # Remove old configuration
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' '/SSH_AUTH_SOCK.*fssh/d' "$CONFIG_FILE"
    else
        sed -i '/SSH_AUTH_SOCK.*fssh/d' "$CONFIG_FILE"
    fi
fi

# Add configuration
echo "" >> "$CONFIG_FILE"
echo "# fssh SSH agent" >> "$CONFIG_FILE"
echo "export SSH_AUTH_SOCK=\"$SOCKET_PATH\"" >> "$CONFIG_FILE"

echo "✓ Added SSH_AUTH_SOCK to $CONFIG_FILE"
echo ""
echo "To apply changes:"
echo "  source $CONFIG_FILE"
echo ""
echo "Or open a new terminal window."
echo ""

# Ask if user wants to apply now
read -p "Apply changes to current shell? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    export SSH_AUTH_SOCK="$SOCKET_PATH"
    echo "✓ Environment variable set for current shell"
    echo ""
    echo "Test with: ssh-add -l"
fi
