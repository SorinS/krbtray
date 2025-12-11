-- gen_token.lua
-- Generate a timestamped token based on the snippet value
-- Attach to snippet entry to generate dynamic tokens
--
-- Context variables available:
--   ctx.value - The snippet value from config (use as base/secret)
--   ctx.name  - Display name of the entry
--   ctx.index - Index number (as string)
--
-- Set 'result' global to specify what gets copied to clipboard

local secret = ctx.value
if secret == "" then
    secret = "default-secret"
end

-- Create timestamp
local timestamp = os.time()
local expires = timestamp + 3600  -- 1 hour from now

-- Build a simple token format: secret-timestamp-expiry
local token = string.format("%s-%d-%d", secret, timestamp, expires)

-- Set result to be copied to clipboard
result = token

ktray.set_status("Token generated (expires in 1h)")
ktray.log("Generated token for: " .. ctx.name)
ktray.log("Token: " .. token)