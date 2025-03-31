-- Neovim plugin for Cobutler AI code completion
-- Similar to Github Copilot but using Cobutler API
if vim.g.loaded_cobutler then
  return
end
vim.g.loaded_cobutler = true

-- Plugin commands
vim.api.nvim_create_user_command('CobutlerEnable', function()
  require('cobutler').enable()
end, {})

vim.api.nvim_create_user_command('CobutlerDisable', function()
  require('cobutler').disable()
end, {})

vim.api.nvim_create_user_command('CobutlerToggle', function()
  require('cobutler').toggle()
end, {})

vim.api.nvim_create_user_command('CobutlerStatus', function()
  require('cobutler').status()
end, {})

-- Auto setup
vim.defer_fn(function()
  require('cobutler').setup()
end, 0) 