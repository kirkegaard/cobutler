local M = {}
local api = require('cobutler.api')
local config = require('cobutler.config')
local util = require('cobutler.util')

local ns_id = vim.api.nvim_create_namespace('cobutler')
local current_suggestion = nil
local current_buffer = nil
local active_extmark = nil
local current_line = nil

-- Check if the current buffer should be processed
local function should_process_buffer(bufnr)
  bufnr = bufnr or vim.api.nvim_get_current_buf()
  
  -- Check if the buffer is valid
  if not vim.api.nvim_buf_is_valid(bufnr) then
    return false
  end
  
  -- Get the buffer filetype
  local filetype = vim.bo[bufnr].filetype
  
  -- Check if the filetype is excluded
  for _, excluded in ipairs(config.options.filetypes.exclude) do
    if filetype == excluded then
      return false
    end
  end
  
  return true
end

-- Clear current suggestion
local function clear_suggestion()
  if active_extmark and current_buffer and vim.api.nvim_buf_is_valid(current_buffer) then
    vim.api.nvim_buf_del_extmark(current_buffer, ns_id, active_extmark)
    active_extmark = nil
  end
  
  current_suggestion = nil
  current_buffer = nil
  current_line = nil
end

-- Display suggestion as virtual text
local function display_suggestion(bufnr, line, suggestion_text)
  if util.is_empty(suggestion_text) then
    return
  end
  
  -- Clear any existing suggestion
  clear_suggestion()
  
  -- Store current buffer and suggestion
  current_buffer = bufnr
  current_suggestion = suggestion_text
  current_line = line
  
  -- Split suggestion into lines
  local lines = vim.split(suggestion_text, "\n")
  
  -- Only display up to max_lines
  local display_lines = {}
  for i = 1, math.min(#lines, config.options.max_lines) do
    table.insert(display_lines, lines[i])
  end
  
  -- Create the extmark with virtual text
  active_extmark = vim.api.nvim_buf_set_extmark(bufnr, ns_id, line, 0, {
    virt_text_pos = "eol",
    virt_text = {{
      config.options.virtual_text.prefix .. display_lines[1],
      config.options.virtual_text.highlight
    }},
    virt_lines = #display_lines > 1 and vim.tbl_map(function(l)
      return {{l, config.options.virtual_text.highlight}}
    end, vim.list_slice(display_lines, 2)) or nil,
  })
end

-- Get context from the current position with code-specific enhancements
local function get_context(bufnr, row, col)
  bufnr = bufnr or vim.api.nvim_get_current_buf()
  
  -- Get filetype info for better code understanding
  local filetype = vim.bo[bufnr].filetype
  
  -- Adjust context window based on code structure
  local context_lines = 15  -- Increased from 10 for better code context
  
  -- If inside a function or block, try to capture the whole structure
  local start_row = row
  local found_block_start = false
  
  -- Look up to find function or block start
  for i = row, math.max(0, row - 25), -1 do
    local line = vim.api.nvim_buf_get_lines(bufnr, i, i + 1, true)[1] or ""
    local trimmed = line:match("^%s*(.-)%s*$")
    
    -- Check for function definitions, class definitions, or block starts
    if trimmed:match("^function") or
       trimmed:match("^class") or
       trimmed:match("^def ") or
       trimmed:match("^sub ") or
       trimmed:match("^impl") or
       trimmed:match("{%s*$") or
       trimmed:match("^[%w_]+%s*%(.-%)%s*{?%s*$") then
      start_row = i
      found_block_start = true
      break
    end
  end
  
  -- If we didn't find a block start, use increased context window
  if not found_block_start then
    start_row = math.max(0, row - context_lines)
  end
  
  -- Get the previous lines for context
  local prev_lines = vim.api.nvim_buf_get_lines(bufnr, start_row, row, true)
  
  -- Get the current line
  local curr_line = vim.api.nvim_buf_get_lines(bufnr, row, row + 1, true)[1] or ""
  
  -- Get a few lines after the cursor for more context
  local next_lines = {}
  if row + 5 < vim.api.nvim_buf_line_count(bufnr) then
    next_lines = vim.api.nvim_buf_get_lines(bufnr, row + 1, row + 5, true)
  end
  
  -- Build context with special markers
  local context = ""
  
  -- Add filetype hint
  context = "// FILETYPE: " .. filetype .. "\n"
  
  -- Add prev lines
  if #prev_lines > 0 then
    context = context .. table.concat(prev_lines, "\n") .. "\n"
  end
  
  -- Add current line up to cursor
  context = context .. curr_line:sub(1, col)
  
  -- Optionally include current line after cursor and next few lines as a hint
  local hint = ""
  if col < #curr_line then
    hint = hint .. "// AFTER CURSOR: " .. curr_line:sub(col + 1) .. "\n"
  end
  
  -- Add next few lines as hint
  if #next_lines > 0 then
    hint = hint .. "// CONTEXT AFTER: \n// " .. table.concat(next_lines, "\n// ") .. "\n"
  end
  
  -- Combine context plus hint
  if hint ~= "" then
    context = context .. "\n" .. hint
  end
  
  return context
end

-- Regular function to request suggestions
local function request_suggestion()
  if not should_process_buffer() then
    clear_suggestion()
    return
  end
  
  -- Get cursor position
  local bufnr = vim.api.nvim_get_current_buf()
  local row, col = unpack(vim.api.nvim_win_get_cursor(0))
  row = row - 1 -- Convert to 0-based indexing
  
  -- Get context
  local context = get_context(bufnr, row, col)
  
  -- Request completion
  api.get_completion(context, function(completion, err)
    if err or not completion then
      return
    end
    
    -- Display the suggestion
    display_suggestion(bufnr, row, completion)
  end)
end

-- Create a debounced version of the request function
local request_suggestion_debounced = util.debounce(request_suggestion, config.options.debounce_ms)

-- Accept the current suggestion
function M.accept_suggestion()
  if util.is_empty(current_suggestion) or 
     not current_buffer or 
     not vim.api.nvim_buf_is_valid(current_buffer) then
    return false
  end
  
  -- Save needed variables before clearing
  local suggestion_to_insert = current_suggestion
  local buffer_to_modify = current_buffer
  local original_context = get_context(current_buffer, current_line, 0)
  
  -- Clear the suggestion before modifying buffer
  clear_suggestion()
  
  -- Insert the suggestion at cursor position
  vim.schedule(function()
    -- Check if buffer is still valid
    if not vim.api.nvim_buf_is_valid(buffer_to_modify) then
      return
    end
    
    -- Get cursor position
    local cursor_pos = vim.api.nvim_win_get_cursor(0)
    local row = cursor_pos[1] - 1  -- 0-indexed
    local col = cursor_pos[2]      -- 0-indexed
    
    -- Get the current line
    local line = vim.api.nvim_buf_get_lines(buffer_to_modify, row, row + 1, false)[1] or ""
    
    -- Create the new line with the suggestion inserted at cursor position
    local new_line = line:sub(1, col) .. suggestion_to_insert .. line:sub(col + 1)
    
    -- Replace the current line with the new one
    vim.api.nvim_buf_set_lines(buffer_to_modify, row, row + 1, false, {new_line})
    
    -- Move cursor to end of inserted text
    vim.api.nvim_win_set_cursor(0, {row + 1, col + #suggestion_to_insert})
    
    -- Learn from what was accepted, passing the original context
    api.learn(suggestion_to_insert, original_context)
  end)
  
  return true
end

-- Check if we have a suggestion to accept
function M.has_suggestion()
  return not util.is_empty(current_suggestion) and 
         current_buffer and 
         vim.api.nvim_buf_is_valid(current_buffer)
end

-- Setup autocommands and keymap handlers
local function setup_autocmds()
  local group = vim.api.nvim_create_augroup('cobutler', { clear = true })
  
  -- Start suggestion timer on cursor movement or text change
  vim.api.nvim_create_autocmd({ 'CursorMovedI', 'TextChangedI' }, {
    group = group,
    callback = function()
      request_suggestion_debounced()
    end
  })
  
  -- Clear suggestions on mode change
  vim.api.nvim_create_autocmd('ModeChanged', {
    group = group,
    pattern = '*:n',
    callback = function()
      clear_suggestion()
    end
  })
  
  -- Setup keymaps for suggestion handling
  -- Use a regular keymap for accept, not an expr mapping
  vim.keymap.set('i', config.options.keymaps.accept, function()
    if M.has_suggestion() then
      M.accept_suggestion()
    else
      -- Pass through the tab key
      local tab = vim.api.nvim_replace_termcodes("<Tab>", true, false, true)
      vim.api.nvim_feedkeys(tab, 'n', false)
    end
  end)
  
  -- Dismiss keymap
  vim.keymap.set('i', config.options.keymaps.dismiss, function()
    if current_suggestion then
      clear_suggestion()
    else
      -- Pass through the key
      local key = vim.api.nvim_replace_termcodes(config.options.keymaps.dismiss, true, false, true)
      vim.api.nvim_feedkeys(key, 'n', false)
    end
  end)
end

-- Start the suggestion engine
function M.start()
  setup_autocmds()
end

-- Stop the suggestion engine
function M.stop()
  -- Clear any existing suggestion
  clear_suggestion()
  
  -- Clear the autocommand group
  vim.api.nvim_clear_autocmds({ group = 'cobutler' })
end

return M 