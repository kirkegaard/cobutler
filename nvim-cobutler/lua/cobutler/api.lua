local M = {}
local curl = require('plenary.curl')
local config = require('cobutler.config')
local util = require('cobutler.util')

-- Validate that the Cobutler API is accessible
function M.setup()
  vim.schedule(function()
    M.check_connection()
  end)
end

-- Check connection to the API
function M.check_connection()
  local ok, result = pcall(function()
    return curl.post(config.options.api_url .. "/predict", {
      headers = {
        content_type = "application/json",
      },
      body = vim.fn.json_encode({ 
        text = "test connection",
        max_words = config.options.max_reply_length,
        precision = config.options.precision_rating
      }),
      timeout = 3000,
    })
  end)

  if not ok or (result and result.status >= 400) then
    vim.notify("Cobutler: Failed to connect to API server at " .. config.options.api_url, vim.log.levels.WARN)
    return false
  end
  
  return true
end

-- Get a completion for the given context
function M.get_completion(context, callback)
  -- Sanitize the context
  local sanitized_context = util.sanitize_text(context)
  
  vim.schedule(function()
    local ok, result = pcall(function()
      return curl.post(config.options.api_url .. "/predict", {
        headers = {
          content_type = "application/json",
        },
        body = vim.fn.json_encode({ 
          text = sanitized_context,
          max_words = config.options.max_reply_length,
          precision = config.options.precision_rating,
          use_cache = config.options.use_cache,
          debug = config.options.debug
        }),
        timeout = 5000,
      })
    end)

    if not ok or (result and result.status >= 400) then
      vim.schedule(function()
        callback(nil, "Failed to get completion from API")
      end)
      return
    end

    local response = vim.fn.json_decode(result.body)
    
    if not response or not response.reply then
      vim.schedule(function()
        callback(nil, "Invalid response from API")
      end)
      return
    end

    vim.schedule(function()
      callback(response.reply)
    end)
  end)
end

-- Learn from text
function M.learn(text, context)
  -- Sanitize the text and context
  local sanitized_text = util.sanitize_text(text)
  local sanitized_context = context and util.sanitize_text(context) or ""
  
  vim.schedule(function()
    local ok, result = pcall(function()
      return curl.post(config.options.api_url .. "/learn", {
        headers = {
          content_type = "application/json",
        },
        body = vim.fn.json_encode({ 
          text = sanitized_text,
          context = sanitized_context,
          debug = config.options.debug
        }),
        timeout = 5000,
      })
    end)

    if not ok or (result and result.status >= 400) then
      vim.notify("Failed to send learning data to API", vim.log.levels.WARN)
    end
  end)
end

return M 