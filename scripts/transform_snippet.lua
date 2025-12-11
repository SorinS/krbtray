-- transform_snippet.lua
-- Transform the snippet value before copying to clipboard
-- Applies different transformations based on snippet name
--
-- Context variables available:
--   ctx.value - The snippet value from config
--   ctx.name  - Display name of the entry
--   ctx.index - Index number (as string)
--
-- Set 'result' global to specify what gets copied to clipboard

local value = ctx.value

ktray.log("Transforming snippet: " .. ctx.name)
ktray.log("Original value: " .. value)

-- Apply transformation based on snippet name keywords
if ctx.name:lower():find("base64") then
    -- Hex encoding (simple substitute for base64)
    result = value:gsub(".", function(c)
        return string.format("%02x", string.byte(c))
    end)
    ktray.set_status("Hex encoded: " .. ctx.name)

elseif ctx.name:lower():find("json") then
    -- Wrap in JSON object
    local escaped = value:gsub('"', '\\"'):gsub("\n", "\\n")
    result = '{"value": "' .. escaped .. '"}'
    ktray.set_status("JSON wrapped: " .. ctx.name)

elseif ctx.name:lower():find("header") then
    -- Format as HTTP header
    result = "X-Custom-Header: " .. value
    ktray.set_status("Header formatted: " .. ctx.name)

elseif ctx.name:lower():find("upper") then
    -- Convert to uppercase
    result = value:upper()
    ktray.set_status("Uppercased: " .. ctx.name)

elseif ctx.name:lower():find("lower") then
    -- Convert to lowercase
    result = value:lower()
    ktray.set_status("Lowercased: " .. ctx.name)

elseif ctx.name:lower():find("trim") then
    -- Trim whitespace
    result = value:match("^%s*(.-)%s*$")
    ktray.set_status("Trimmed: " .. ctx.name)

else
    -- Default: return as-is
    result = value
    ktray.set_status("Copied: " .. ctx.name)
end

ktray.log("Transformed value: " .. result)