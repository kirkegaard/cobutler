local M = {}

-- Default configuration options
M.defaults = {
  -- API settings
  api_url = "http://localhost:8080",
  max_reply_length = 5, -- Maximum number of words in the reply
  precision_rating = 0.7, -- Controls precision of responses (0.0-1.0, higher = more focused)
  use_cache = false, -- Whether to use token caching (disable to avoid repetition)
  debug = false, -- Enable debug logging on the server side
  
  -- Plugin behavior
  auto_enable = true,
  debounce_ms = 300,
  max_suggestion_length = 60, -- Maximum length of suggestion in characters
  
  -- Suggestion settings
  max_lines = 3,
  suggestion_delay_ms = 100,
  
  -- UI settings
  virtual_text = {
    highlight = "Comment",
    prefix = "» ", -- Prefix for the suggestion
  },
  
  -- Keymaps
  keymaps = {
    accept = "<Tab>", -- Accept suggestion
    dismiss = "<S-Tab>", -- Dismiss suggestion
    next_line = "<C-n>",
    prev_line = "<C-p>",
  },
  
  -- Filetype specific settings
  filetypes = {
    -- Exclude specific filetypes
    exclude = {
      "TelescopePrompt",
      "NvimTree",
      "neo-tree",
      "dashboard",
      "alpha",
      "lazy",
      "mason",
    },
  },
}

-- Active configuration options
M.options = {}

-- Setup configuration
function M.setup(opts)
  M.options = vim.tbl_deep_extend("force", {}, M.defaults, opts or {})
  
  -- Clamp precision rating to valid range
  if M.options.precision_rating < 0 then M.options.precision_rating = 0 end
  if M.options.precision_rating > 1 then M.options.precision_rating = 1 end
end

return M 