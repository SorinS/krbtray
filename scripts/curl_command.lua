-- curl_command.lua
-- Generate a curl command with Kerberos authentication
-- Copies a ready-to-use curl command to clipboard
--
-- Context variables available:
--   ctx.url   - The URL from the entry configuration
--   ctx.name  - Display name of the entry
--   ctx.index - Index number (as string)
--
-- Set 'result' global to specify what gets copied to clipboard
-- (This script can be attached to either URL or snippet entries)

ktray.set_status("Generating curl command...")

-- Get the current Kerberos token
local token, err = ktray.get_token()
if not token then
    -- Generate curl with --negotiate flag instead
    result = string.format(
        'curl --negotiate -u : -H "Accept: application/json" "%s"',
        ctx.url
    )
    ktray.set_status("curl with --negotiate copied")
    ktray.log("No token available, using --negotiate flag")
    return
end

-- Generate curl command with explicit Authorization header
result = string.format(
    'curl -H "Authorization: Negotiate %s" -H "Accept: application/json" "%s"',
    token,
    ctx.url
)

ktray.set_status("curl command copied")
ktray.log("Generated curl for: " .. ctx.url)