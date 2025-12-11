-- api_auth.lua
-- Fetch API data with Kerberos SPNEGO authentication
-- Attach to URL entry to make authenticated requests
--
-- Context variables available:
--   ctx.url   - The URL from the entry configuration
--   ctx.name  - Display name of the entry
--   ctx.index - Index number (as string)

ktray.set_status("Fetching " .. ctx.name .. "...")

-- Get the current Kerberos token
local token, err = ktray.get_token()
if not token then
    ktray.set_status("Error: No Kerberos token available")
    ktray.notify("Auth Failed", "Please refresh your Kerberos ticket")
    return
end

ktray.log("Got Kerberos token, making request to: " .. ctx.url)

-- Make authenticated request
local headers = {
    ["Authorization"] = "Negotiate " .. token,
    ["Accept"] = "application/json"
}

local body, err = ktray.http_get(ctx.url, headers)
if err then
    ktray.set_status("API error: " .. err)
    ktray.log("HTTP GET failed: " .. err)
    return
end

-- Copy response to clipboard
local ok, copy_err = ktray.copy(body)
if ok then
    ktray.set_status("API response copied (" .. #body .. " bytes)")
    ktray.log("Response copied to clipboard")
else
    ktray.set_status("Copy failed: " .. (copy_err or "unknown"))
end