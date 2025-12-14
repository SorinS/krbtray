-- bam_rsa_auth.lua
-- BAM authentication with RSA SecurID two-factor authentication
--
-- Flow:
-- 1. Request BAM token (triggers RSA form if privileged)
-- 2. Detect if RSA authentication is required
-- 3. Prompt user for RSA PIN
-- 4. POST RSA PIN back to same URL
-- 5. Capture BAM JWT token from response and cache it
--
-- Context variables:
--   ctx.url   - The BAM endpoint URL
--   ctx.name  - Display name of the entry

local CACHE_KEY = "bam_token"
local CACHE_TTL = 3600  -- 1 hour

-- Check cache first
local cached_token, found = ktray.cache_get(CACHE_KEY)
if found and cached_token ~= "" then
    ktray.log("Using cached BAM token")
    ktray.copy(cached_token)
    ktray.set_status("BAM token copied (cached)")
    result = cached_token
    return
end

ktray.set_status("Requesting BAM token...")

-- Get Kerberos token for BAM
local krb_token, err = ktray.get_token("BAM")
if not krb_token then
    ktray.set_status("Error: " .. (err or "No Kerberos token"))
    ktray.notify("Auth Failed", "Could not get Kerberos token for BAM")
    return
end

ktray.log("Got Kerberos token, requesting BAM...")

-- Initial request with Kerberos auth
local headers = {
    ["Authorization"] = "Negotiate " .. krb_token,
    ["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
}

local response, err = ktray.http_get(ctx.url, headers, 30, true)
if err then
    ktray.set_status("Request failed: " .. err)
    ktray.log("HTTP GET failed: " .. err)
    return
end

ktray.log("Got response, length: " .. #response)

-- Check if response is already a JWT token (starts with eyJ)
if string.sub(response, 1, 3) == "eyJ" then
    ktray.log("Got JWT token directly (no RSA required)")
    ktray.cache_set(CACHE_KEY, response, CACHE_TTL)
    ktray.copy(response)
    ktray.set_status("BAM token copied")
    result = response
    return
end

-- Check if we got an RSA form (look for RSA-related elements)
local rsa_elements = ktray.html_find(response, "#RSAAuthOption, #oatCode, input[name='OATCode']")
if #rsa_elements == 0 then
    ktray.log("No RSA form detected and no token found")
    ktray.set_status("Unexpected response")
    ktray.log("Response preview: " .. string.sub(response, 1, 500))
    return
end

ktray.log("RSA authentication required")

-- Prompt user for RSA PIN
local rsa_pin, ok = ktray.prompt_secret("RSA Authentication", "Enter your RSA SecurID code:")
if not ok or rsa_pin == "" then
    ktray.set_status("Cancelled")
    return
end

ktray.set_status("Submitting RSA code...")

-- Extract hidden fields from the response
local auth_token = ktray.html_attr(response, "input#authenticationToken", "value") or ""
local token_required = ktray.html_attr(response, "input#TokenRequired", "value") or "false"

ktray.info("Hidden fields - authenticationToken: " .. string.sub(auth_token, 1, 20) .. "..., TokenRequired: " .. token_required)

-- POST RSA PIN back to same URL (form-urlencoded with hidden fields)
local post_body = "OATCode=" .. rsa_pin .. "&authenticationToken=" .. auth_token .. "&TokenRequired=" .. token_required

local post_headers = {
    ["Authorization"] = "Negotiate " .. krb_token,
    ["Content-Type"] = "application/x-www-form-urlencoded",
    ["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
}

local rsa_response, err = ktray.http_post(ctx.url, post_body, post_headers, 30, true)
if err then
    ktray.set_status("RSA submit failed: " .. err)
    ktray.log("RSA POST failed: " .. err)
    return
end

ktray.log("RSA response length: " .. #rsa_response)

-- Extract the BAM JWT token from response
local bam_token = nil

-- Check if response is JWT directly (most likely case)
if string.sub(rsa_response, 1, 3) == "eyJ" then
    bam_token = string.match(rsa_response, "^[^\n\r]+")  -- First line only, trim whitespace
    ktray.log("Got JWT token from RSA response")
end

if not bam_token or bam_token == "" then
    ktray.set_status("Failed to extract BAM token")
    ktray.log("Could not find JWT in response")
    ktray.log("Response preview: " .. string.sub(rsa_response, 1, 500))
    return
end

-- Cache and return the token
ktray.cache_set(CACHE_KEY, bam_token, CACHE_TTL)
ktray.copy(bam_token)
ktray.set_status("BAM token copied and cached")
ktray.log("BAM token obtained and cached")

result = bam_token
