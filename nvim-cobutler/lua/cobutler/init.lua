local M = {}
local api = require('cobutler.api')
local config = require('cobutler.config')
local suggestion = require('cobutler.suggestion')

local enabled = false
local initialized = false

-- Initialize the plugin
function M.setup(opts)
  if initialized then
    return
  end
  
  config.setup(opts or {})
  api.setup()
  initialized = true
  
  if config.options.auto_enable then
    M.enable()
  end
end

-- Enable the plugin
function M.enable()
  if not initialized then
    M.setup()
  end
  
  if enabled then
    return
  end
  
  enabled = true
  suggestion.start()
  
  vim.notify("Cobutler enabled", vim.log.levels.INFO)
end

-- Disable the plugin
function M.disable()
  if not enabled then
    return
  end
  
  enabled = false
  suggestion.stop()
  
  vim.notify("Cobutler disabled", vim.log.levels.INFO)
end

-- Toggle the plugin
function M.toggle()
  if enabled then
    M.disable()
  else
    M.enable()
  end
end

-- Show plugin status
function M.status()
  if not initialized then
    vim.notify("Cobutler: not initialized", vim.log.levels.INFO)
    return
  end
  
  local status = enabled and "enabled" or "disabled"
  vim.notify("Cobutler: " .. status, vim.log.levels.INFO)
end

-- Check if plugin is enabled
function M.is_enabled()
  return enabled
end

return M 