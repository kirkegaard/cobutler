#!/bin/bash

# nvim-cobutler installation script

# Find Neovim config directory
if [ -d "$HOME/.config/nvim" ]; then
    CONFIG_DIR="$HOME/.config/nvim"
elif [ -d "$HOME/.local/share/nvim" ]; then
    CONFIG_DIR="$HOME/.local/share/nvim"
elif [ -d "$HOME/AppData/Local/nvim" ]; then
    CONFIG_DIR="$HOME/AppData/Local/nvim"
else
    echo "Could not find Neovim config directory."
    echo "Please manually install the plugin to your Neovim config directory."
    exit 1
fi

# Plugin directory to install to
PLUGIN_DIR="$CONFIG_DIR/pack/plugins/start/nvim-cobutler"

echo "Installing nvim-cobutler to $PLUGIN_DIR..."

# Create plugin directory if it doesn't exist
mkdir -p "$PLUGIN_DIR"

# Copy all files to the plugin directory
cp -r lua plugin doc LICENSE README.md "$PLUGIN_DIR"

echo "Installation complete!"
echo ""
echo "NOTE: This plugin requires plenary.nvim to be installed."
echo "If you don't have it installed, add it with your plugin manager or manually."
echo ""
echo "Don't forget to start the Cobutler API server before using the plugin:"
echo "  cd /path/to/cobutler"
echo "  go run cmd/cobutler/main.go"
echo ""
echo "To customize the plugin, add this to your Neovim config:"
echo ""
echo "lua << EOF"
echo "require('cobutler').setup({" 
echo "  -- Your config options here"
echo "})"
echo "EOF" 