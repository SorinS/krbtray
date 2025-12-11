-- env_url.lua
-- Select URL based on environment variable
-- Opens dev/staging/prod URL variants based on APP_ENV
--
-- Context variables available:
--   ctx.url   - Base URL from the entry configuration
--   ctx.name  - Display name of the entry
--   ctx.index - Index number (as string)
--
-- Set APP_ENV environment variable to: dev, development, staging, or production

local env = ktray.env("APP_ENV") or "production"
ktray.log("Current environment: " .. env)

-- Build URL based on environment
local base_url = ctx.url
local final_url

if env == "development" or env == "dev" then
    -- Insert "dev." after protocol
    final_url = base_url:gsub("://", "://dev.")
    ktray.set_status("Opening DEV: " .. ctx.name)
elseif env == "staging" or env == "stage" then
    -- Insert "staging." after protocol
    final_url = base_url:gsub("://", "://staging.")
    ktray.set_status("Opening STAGING: " .. ctx.name)
else
    -- Production - use URL as-is
    final_url = base_url
    ktray.set_status("Opening PROD: " .. ctx.name)
end

ktray.log("Base URL: " .. base_url)
ktray.log("Final URL: " .. final_url)

local ok, err = ktray.open_url(final_url)
if not ok then
    ktray.set_status("Failed to open: " .. (err or "unknown error"))
end