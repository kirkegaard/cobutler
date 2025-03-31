local M = {}

-- Check if a value is empty (nil or empty string/table)
function M.is_empty(value)
  if value == nil then
    return true
  end
  
  if type(value) == "string" then
    return value == ""
  end
  
  if type(value) == "table" then
    return vim.tbl_isempty(value)
  end
  
  return false
end

-- Safely get a nested table field
function M.get_nested(tbl, path, default)
  local current = tbl
  
  for _, key in ipairs(path) do
    if type(current) ~= "table" then
      return default
    end
    current = current[key]
    if current == nil then
      return default
    end
  end
  
  return current
end

-- Debounce a function call
function M.debounce(fn, ms)
  -- Ensure ms is a number with a default value
  ms = type(ms) == "number" and ms or 300
  
  local timer = nil
  return function(...)
    -- Store args for later
    local args = {...}
    
    -- Cancel previous timer if it exists
    if timer then
      timer:stop()
      timer = nil
    end
    
    -- Create a new timer
    timer = vim.loop.new_timer()
    
    -- Start the timer with the specified delay
    timer:start(ms, 0, vim.schedule_wrap(function()
      -- Stop and clear the timer
      if timer then
        timer:stop()
        timer:close()
        timer = nil
      end
      
      -- Call the function with original args
      fn(unpack(args))
    end))
  end
end

-- Sanitize text for API requests
function M.sanitize_text(text)
  if not text then
    return ""
  end
  
  -- Remove control characters
  text = text:gsub("%c", " ")
  
  -- Truncate long text to prevent large API requests
  if #text > 1000 then
    text = text:sub(#text - 1000)
  end
  
  return text
end

-- Get filetype based on file extension
function M.get_filetype_from_extension(filename)
  local extension = filename:match("%.([^.]+)$")
  
  if not extension then
    return nil
  end
  
  local extension_to_filetype = {
    lua = "lua",
    py = "python",
    js = "javascript",
    ts = "typescript",
    jsx = "javascriptreact",
    tsx = "typescriptreact",
    go = "go",
    rs = "rust",
    c = "c",
    cpp = "cpp",
    h = "c",
    hpp = "cpp",
    java = "java",
    rb = "ruby",
    php = "php",
    html = "html",
    css = "css",
    json = "json",
    md = "markdown",
  }
  
  return extension_to_filetype[extension:lower()]
end

return M 