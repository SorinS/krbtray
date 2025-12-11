-- pre_ssh.lua
-- Run setup commands before opening SSH connection
-- Then open the terminal with the SSH command
--
-- Context variables available:
--   ctx.command  - SSH command (e.g., "ssh user@host")
--   ctx.terminal - Terminal template with {cmd} placeholder
--   ctx.name     - Display name of the entry
--   ctx.index    - Index number (as string)

ktray.log("Pre-SSH setup for: " .. ctx.name)
ktray.set_status("Preparing SSH: " .. ctx.name)

-- Add SSH keys to agent (ignore errors if already added or key doesn't exist)
ktray.shell("ssh-add ~/.ssh/id_rsa 2>/dev/null")
ktray.shell("ssh-add ~/.ssh/id_ed25519 2>/dev/null")

-- Extract host from command for connectivity check
local host = ctx.command:match("@([%w%.%-]+)")
if host then
    ktray.log("Checking connectivity to: " .. host)

    -- Quick ping test (1 packet, 2 second timeout)
    -- Note: -W is timeout on Linux, -t on macOS
    local output, err = ktray.shell("ping -c 1 -t 2 " .. host .. " 2>&1 || ping -c 1 -W 2 " .. host .. " 2>&1")
    if err and not output:find("1 packets received") and not output:find("1 received") then
        ktray.set_status("Warning: " .. host .. " may be unreachable")
        ktray.sleep(1500)  -- Show warning briefly
    end
end

-- Now open the terminal with SSH command
-- Replace {cmd} placeholder in terminal template
local terminal_cmd = ctx.terminal:gsub("{cmd}", ctx.command)
ktray.log("Executing: " .. terminal_cmd)

local output, err = ktray.shell(terminal_cmd)
if err then
    ktray.set_status("SSH failed: " .. err)
    ktray.log("Terminal launch error: " .. err)
else
    ktray.set_status("SSH: " .. ctx.name)
end