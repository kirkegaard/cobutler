# nvim-cobutler

A Neovim plugin for AI code suggestions using Cobutler. Similar to GitHub Copilot, but using your local Cobutler API.

## Features

- AI-powered code suggestions as you type
- Non-intrusive UI with virtual text
- Easy to accept or dismiss suggestions
- Learns from your accepted suggestions

## Requirements

- Neovim >= 0.8.0
- [plenary.nvim](https://github.com/nvim-lua/plenary.nvim) for HTTP requests
- Cobutler API server running

## Installation

### Using [packer.nvim](https://github.com/wbthomason/packer.nvim)

```lua
use {
  'kirkegaard/nvim-cobutler',
  requires = { 'nvim-lua/plenary.nvim' }
}
```

### Using [lazy.nvim](https://github.com/folke/lazy.nvim)

```lua
{
  'kirkegaard/nvim-cobutler',
  dependencies = { 'nvim-lua/plenary.nvim' }
}
```

## Setup

```lua
require('cobutler').setup({
  -- API settings
  api_url = "http://localhost:8080", -- URL of your Cobutler API
  
  -- Plugin behavior
  auto_enable = true, -- Enable on startup
  debounce_ms = 300, -- Delay before requesting suggestion
  
  -- Suggestion settings
  max_lines = 5, -- Maximum number of lines to display
  
  -- Keymaps
  keymaps = {
    accept = "<Tab>", -- Accept the current suggestion
    dismiss = "<C-]>", -- Dismiss the current suggestion
    next_line = "<C-n>", -- Go to next suggestion line (future feature)
    prev_line = "<C-p>", -- Go to previous suggestion line (future feature)
  },
})
```

## Usage

Once the plugin is installed and configured, it will automatically display suggestions as you type.

### Commands

- `:CobutlerEnable` - Enable the plugin
- `:CobutlerDisable` - Disable the plugin
- `:CobutlerToggle` - Toggle the plugin on/off
- `:CobutlerStatus` - Show the current status of the plugin

### Keymaps

- `<Tab>` (or your configured accept key) - Accept the current suggestion
- `<C-]>` (or your configured dismiss key) - Dismiss the current suggestion

## API Server

Make sure your Cobutler API server is running before using this plugin:

```bash
# Navigate to your Cobutler directory
cd path/to/cobutler

# Run the server
go run cmd/cobutler/main.go
```

## License

MIT 