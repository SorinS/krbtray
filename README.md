# krb5tray

A cross-platform system tray application for obtaining Kerberos service tickets and generating SPNEGO tokens for HTTP authentication.

## Platform Support

| Platform | Implementation | Credential Source | Build |
|----------|---------------|-------------------|-------|
| macOS 11+ | GSS API (SPNEGO) | System credential cache via GSSCred | Native or cross-compile |
| Windows | SSPI (Negotiate) | LSA credential cache | Native or cross-compile |
| Linux | gokrb5 (SPNEGO) | File-based ccache | Native only (requires CGO) |

## Prerequisites

### macOS
- macOS 11 (Big Sur) or later
- Valid Kerberos ticket (obtained via `kinit` or domain login)
- Xcode Command Line Tools (for building)

### Windows
- Domain-joined machine or valid Kerberos ticket
- Credentials in LSA cache
- Go 1.19+ (for building)

### Linux
- Valid ccache file (obtained via `kinit`)
- `/etc/krb5.conf` configured (or set `KRB5_CONFIG` environment variable)
- GTK3 development libraries (for systray support)
- Go 1.19+ with CGO enabled (for building)

**Linux GTK dependencies:**

```bash
# Ubuntu/Debian
sudo apt-get install libgtk-3-dev libappindicator3-dev

# Fedora/RHEL
sudo dnf install gtk3-devel libappindicator-gtk3-devel

# Arch Linux
sudo pacman -S gtk3 libappindicator-gtk3
```

## Building

### macOS (native)

```bash
go build -o krb5tray .
```

### macOS (cross-compile for Windows)

```bash
GOOS=windows GOARCH=amd64 go build -o krb5tray.exe .
```

### Windows (native)

```cmd
go build -o krb5tray.exe .
```

### Linux (native only - requires CGO)

```bash
# Ensure GTK dependencies are installed first
CGO_ENABLED=1 go build -o krb5tray .
```

**Note:** Linux cannot be cross-compiled from other platforms due to GTK/CGO dependencies.

## Configuration

### Environment Variables

| Variable | Description | Platform |
|----------|-------------|----------|
| `KRB5_SPN` | Default Service Principal Name (e.g., `HTTP/server.example.com`) | All |
| `KRB5CCNAME` | Path to credential cache file | Linux |
| `KRB5_CONFIG` | Path to krb5.conf (default: `/etc/krb5.conf`) | Linux |

### Configuration File

krb5tray uses a JSON configuration file located at `~/.config/ktray/ktray.json`. The file supports the following sections:

```json
{
  "spns": [
    {"name": "Production API", "spn": "HTTP/api.example.com@REALM.COM"},
    {"name": "Staging API", "spn": "HTTP/api-staging.example.com@REALM.COM"}
  ],
  "secrets": [
    {
      "name": "Database Credentials",
      "auth_url": "https://vault.example.com/auth",
      "role_name": "db-reader",
      "role_type": "approle",
      "rotate_url": "https://vault.example.com/rotate",
      "secret_url": "https://vault.example.com/secret"
    }
  ],
  "urls": [
    {"index": 0, "name": "Jira", "url": "https://jira.example.com"},
    {"index": 1, "name": "Confluence", "url": "https://confluence.example.com"}
  ],
  "snippets": [
    {"index": 1, "name": "Bearer Token", "value": "Bearer abc123..."},
    {"index": 2, "name": "API Key", "value": "x-api-key: secret123"}
  ],
  "ssh": [
    {"index": 0, "name": "Prod Server", "command": "ssh admin@prod.example.com", "terminal": "/Applications/Utilities/Terminal.app/Contents/MacOS/Terminal {cmd}"},
    {"index": 1, "name": "Dev Server", "command": "ssh -A dev@dev.example.com", "terminal": "/usr/bin/gnome-terminal -- {cmd}"}
  ]
}
```

| Section | Description |
|---------|-------------|
| `spns` | Service Principal Names for Kerberos tickets |
| `secrets` | CSM secret configurations |
| `urls` | URL bookmarks that open in browser (use `index` for hotkey access) |
| `snippets` | Text snippets copied to clipboard (use `index` for hotkey access) |
| `ssh` | SSH connections opened in user-defined terminal (use `index` for hotkey access) |

### SSH Terminal Configuration

The `terminal` field in SSH entries is a command template with `{cmd}` as a placeholder for the SSH command. Examples for different terminals:

| Platform | Terminal | Template |
|----------|----------|----------|
| macOS | Terminal.app | `/Applications/Utilities/Terminal.app/Contents/MacOS/Terminal {cmd}` |
| macOS | iTerm2 | `/Applications/iTerm.app/Contents/MacOS/iTerm2 {cmd}` |
| macOS | Alacritty | `/Applications/Alacritty.app/Contents/MacOS/alacritty -e {cmd}` |
| Linux | gnome-terminal | `/usr/bin/gnome-terminal -- {cmd}` |
| Linux | Alacritty | `/usr/bin/alacritty -e {cmd}` |
| Linux | Konsole | `/usr/bin/konsole -e {cmd}` |
| Windows | cmd.exe | `C:\\Windows\\System32\\cmd.exe /k {cmd}` |
| Windows | Windows Terminal | `wt.exe {cmd}` |
| Windows | PowerShell | `powershell.exe -NoExit -Command {cmd}` |

## Lua Scripting

krb5tray supports Lua 5.1 scripting for custom automation via [gopher-lua](https://github.com/yuin/gopher-lua). Scripts are stored in `~/.config/ktray/scripts/` and can be attached to URLs, snippets, and SSH entries.

### Script Location

Scripts must be placed in the scripts directory:
```
~/.config/ktray/
├── ktray.json          # Configuration file
└── scripts/            # Lua scripts directory
    ├── api_auth.lua
    ├── gen_token.lua
    └── pre_ssh.lua
```

### Attaching Scripts to Entries

Add a `script` field to any URL, snippet, or SSH entry. The script filename is relative to the scripts directory:

```json
{
  "urls": [
    {
      "index": 0,
      "name": "API with Auth",
      "url": "https://api.example.com/data",
      "script": "api_auth.lua"
    }
  ],
  "snippets": [
    {
      "index": 1,
      "name": "JWT Token",
      "value": "my-secret-key",
      "script": "gen_jwt.lua"
    }
  ],
  "ssh": [
    {
      "index": 0,
      "name": "Production Server",
      "command": "ssh admin@prod.example.com",
      "terminal": "/Applications/Utilities/Terminal.app/Contents/MacOS/Terminal {cmd}",
      "script": "pre_ssh.lua"
    }
  ]
}
```

### Script Behavior

When an entry with a `script` field is triggered (via menu click or hotkey):

1. The script runs **instead of** the default action
2. Scripts receive context via the `ctx` global table
3. For snippets: set the `result` global to specify clipboard content
4. Scripts can call `ktray.*` functions to perform actions

### Context Variables (`ctx` table)

Each entry type receives different context variables:

**URL entries:**
| Variable | Type | Description |
|----------|------|-------------|
| `ctx.url` | string | The URL from the entry |
| `ctx.name` | string | Display name of the entry |
| `ctx.index` | string | Index number (as string) |

**Snippet entries:**
| Variable | Type | Description |
|----------|------|-------------|
| `ctx.value` | string | The snippet value from config |
| `ctx.name` | string | Display name of the entry |
| `ctx.index` | string | Index number (as string) |

**SSH entries:**
| Variable | Type | Description |
|----------|------|-------------|
| `ctx.command` | string | SSH command (e.g., "ssh user@host") |
| `ctx.terminal` | string | Terminal template with {cmd} placeholder |
| `ctx.name` | string | Display name of the entry |
| `ctx.index` | string | Index number (as string) |

### Returning Values from Scripts

**For snippet scripts:** Set the global `result` variable to specify what gets copied to clipboard:

```lua
-- The result will be copied to clipboard
result = "generated-value-" .. os.time()
```

If `result` is not set or is empty, nothing is copied and the status shows "Script: <name>".

**For URL/SSH scripts:** No return value is used. Use `ktray.*` functions to perform actions.

### Available Functions (`ktray` module)

#### Clipboard Functions

```lua
-- Copy text to clipboard
-- Returns: true on success, or false and error message on failure
local ok, err = ktray.copy("text to copy")
if not ok then
    ktray.log("Copy failed: " .. err)
end

-- Get text from clipboard (currently returns empty string)
local text = ktray.paste()
```

#### Browser Functions

```lua
-- Open URL in default browser
-- Returns: true on success, or false and error message on failure
local ok, err = ktray.open_url("https://example.com")
```

#### HTTP Functions

```lua
-- HTTP GET request
-- Parameters: url (string), headers (table, optional)
-- Returns: body (string) on success, or nil and error message on failure
local body, err = ktray.http_get("https://api.example.com/data")
if err then
    ktray.set_status("GET failed: " .. err)
    return
end

-- HTTP GET with custom headers
local headers = {
    ["Authorization"] = "Bearer token123",
    ["Accept"] = "application/json"
}
local body, err = ktray.http_get("https://api.example.com/data", headers)

-- HTTP POST request
-- Parameters: url (string), body (string), headers (table, optional)
-- Returns: response (string) on success, or nil and error message on failure
local response, err = ktray.http_post(
    "https://api.example.com/submit",
    '{"key": "value"}',
    {["Content-Type"] = "application/json"}
)
```

#### Kerberos Functions

```lua
-- Get the current Kerberos token (base64 encoded)
-- Returns: token (string) on success, or nil and error message if no token
local token, err = ktray.get_token()
if not token then
    ktray.set_status("No Kerberos token: " .. (err or "unknown"))
    return
end

-- Get the current SPN
-- Returns: spn (string), may be empty if not set
local spn = ktray.get_spn()
```

#### Shell Execution Functions

```lua
-- Execute a command with arguments
-- Parameters: command (string), args... (strings)
-- Returns: output (string), and optionally error message
local output, err = ktray.exec("ls", "-la", "/tmp")
if err then
    ktray.log("Command failed: " .. err)
end

-- Execute a shell command (via sh -c on Unix, cmd /c on Windows)
-- Parameters: command (string)
-- Returns: output (string), and optionally error message
local output, err = ktray.shell("echo $HOME && ls -la")
```

#### UI Functions

```lua
-- Set the status line text in the tray menu
ktray.set_status("Operation completed successfully")

-- Show a notification (currently updates status line)
-- Parameters: title (string), message (string, optional)
ktray.notify("Success", "Token copied to clipboard")
ktray.notify("Done")  -- message is optional
```

#### Utility Functions

```lua
-- Pause execution for specified milliseconds
ktray.sleep(1000)  -- sleep for 1 second

-- Get environment variable value
-- Returns: value (string), empty if not set
local home = ktray.env("HOME")
local user = ktray.env("USER")

-- Log message to console (only visible when debug mode is enabled)
ktray.log("Debug: processing request...")
```

### Complete Example Scripts

#### 1. Authenticated API Request (`api_auth.lua`)

Fetches data from an API using Kerberos authentication and copies the response:

```lua
-- api_auth.lua
-- Fetch API data with Kerberos SPNEGO authentication
-- Attach to URL entry to make authenticated requests

ktray.set_status("Fetching " .. ctx.name .. "...")

-- Get the current Kerberos token
local token, err = ktray.get_token()
if not token then
    ktray.set_status("Error: No Kerberos token available")
    ktray.notify("Auth Failed", "Please refresh your Kerberos ticket")
    return
end

-- Make authenticated request
local headers = {
    ["Authorization"] = "Negotiate " .. token,
    ["Accept"] = "application/json"
}

local body, err = ktray.http_get(ctx.url, headers)
if err then
    ktray.set_status("API error: " .. err)
    return
end

-- Copy response to clipboard
local ok, copy_err = ktray.copy(body)
if ok then
    ktray.set_status("API response copied (" .. #body .. " bytes)")
else
    ktray.set_status("Copy failed: " .. (copy_err or "unknown"))
end
```

#### 2. Dynamic JWT Token Generator (`gen_jwt.lua`)

Generates a simple JWT-like token with timestamp:

```lua
-- gen_jwt.lua
-- Generate a timestamped token based on the snippet value
-- The ctx.value contains the base secret from config

local secret = ctx.value
if secret == "" then
    secret = "default-secret"
end

-- Create timestamp
local timestamp = os.time()
local expires = timestamp + 3600  -- 1 hour from now

-- Build a simple token (not a real JWT, just for demo)
local token_data = string.format(
    '{"sub":"user","iat":%d,"exp":%d,"key":"%s"}',
    timestamp,
    expires,
    secret
)

-- Base64-like encoding (simple version)
local b64 = "eyJ0eXAiOiJKV1QifQ."  -- fake header
b64 = b64 .. token_data:gsub(".", function(c)
    return string.format("%02x", string.byte(c))
end)

-- Set result to be copied to clipboard
result = "Bearer " .. b64

ktray.set_status("Token generated (expires in 1h)")
ktray.log("Generated token for: " .. ctx.name)
```

#### 3. Pre-SSH Setup Script (`pre_ssh.lua`)

Performs setup before SSH connection (e.g., add keys, check connectivity):

```lua
-- pre_ssh.lua
-- Run setup commands before opening SSH connection
-- Then open the terminal with the SSH command

ktray.log("Pre-SSH setup for: " .. ctx.name)
ktray.set_status("Preparing SSH: " .. ctx.name)

-- Add SSH key to agent (ignore errors if already added)
ktray.shell("ssh-add ~/.ssh/id_rsa 2>/dev/null")
ktray.shell("ssh-add ~/.ssh/id_ed25519 2>/dev/null")

-- Extract host from command for connectivity check
local host = ctx.command:match("@([%w%.%-]+)")
if host then
    ktray.log("Checking connectivity to: " .. host)

    -- Quick ping test (1 packet, 2 second timeout)
    local output, err = ktray.shell("ping -c 1 -W 2 " .. host .. " 2>&1")
    if err then
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
else
    ktray.set_status("SSH: " .. ctx.name)
end
```

#### 4. URL with Custom Headers (`fetch_with_headers.lua`)

Opens a URL after fetching data with custom headers:

```lua
-- fetch_with_headers.lua
-- Fetch data with custom headers, then open URL in browser

ktray.set_status("Fetching data...")

-- Get some data first (e.g., a session token)
local session_url = ktray.env("SESSION_API") or "https://auth.example.com/session"
local session, err = ktray.http_get(session_url)

if err then
    ktray.log("Could not get session: " .. err)
    -- Continue anyway, just open the URL
end

-- Open the URL in browser
local ok, err = ktray.open_url(ctx.url)
if ok then
    ktray.set_status("Opened: " .. ctx.name)
else
    ktray.set_status("Failed to open: " .. err)
end
```

#### 5. Copy with Transformation (`transform_snippet.lua`)

Transforms snippet value before copying:

```lua
-- transform_snippet.lua
-- Transform the snippet value (e.g., encode, wrap, format)

local value = ctx.value

-- Example transformations based on snippet name
if ctx.name:find("Base64") then
    -- Simple hex encoding as placeholder for base64
    result = value:gsub(".", function(c)
        return string.format("%02x", string.byte(c))
    end)
    ktray.set_status("Encoded: " .. ctx.name)

elseif ctx.name:find("JSON") then
    -- Wrap in JSON
    result = '{"value": "' .. value:gsub('"', '\\"') .. '"}'
    ktray.set_status("JSON wrapped: " .. ctx.name)

elseif ctx.name:find("Header") then
    -- Format as HTTP header
    result = "X-Custom-Header: " .. value
    ktray.set_status("Header formatted: " .. ctx.name)

else
    -- Default: uppercase
    result = value:upper()
    ktray.set_status("Transformed: " .. ctx.name)
end

ktray.log("Original: " .. value)
ktray.log("Result: " .. result)
```

#### 6. Environment-based URL (`env_url.lua`)

Opens different URLs based on environment:

```lua
-- env_url.lua
-- Select URL based on environment variable

local env = ktray.env("APP_ENV") or "production"
ktray.log("Current environment: " .. env)

-- Build URL based on environment
local base_url = ctx.url
local final_url

if env == "development" or env == "dev" then
    final_url = base_url:gsub("://", "://dev.")
    ktray.set_status("Opening DEV: " .. ctx.name)
elseif env == "staging" then
    final_url = base_url:gsub("://", "://staging.")
    ktray.set_status("Opening STAGING: " .. ctx.name)
else
    final_url = base_url
    ktray.set_status("Opening PROD: " .. ctx.name)
end

ktray.log("Final URL: " .. final_url)
ktray.open_url(final_url)
```

### Setting the SPN (Alternative)

You can also set a default SPN via environment variable:

```bash
# macOS/Linux
export KRB5_SPN=HTTP/server.example.com

# Windows (cmd)
set KRB5_SPN=HTTP/server.example.com

# Windows (PowerShell)
$env:KRB5_SPN = "HTTP/server.example.com"
```

## Usage

1. Ensure you have a valid Kerberos ticket:
   ```bash
   # Check existing tickets
   klist

   # Or obtain a new ticket
   kinit user@REALM
   ```

2. Set the target SPN environment variable (see above)

3. Run the application:
   ```bash
   ./krb5tray
   ```

4. Use the system tray menu:
   - **Get Ticket** - Request a service ticket for the configured SPN
   - **Copy HTTP Header** - Copy `Negotiate <token>` to clipboard (for use in HTTP Authorization header)
   - **Copy Token** - Copy just the base64 token to clipboard
   - **Debug Mode** - Toggle debug output
   - **Quit** - Exit the application

## Tray Menu Options

| Menu Item | Description |
|-----------|-------------|
| Status line | Shows current platform, ticket status, or errors |
| Select SPN | Submenu to choose a service principal from config |
| CSM Secrets | Submenu to manage CSM secrets |
| URLs | Submenu to open configured URLs in browser |
| Snippets | Submenu to copy text snippets to clipboard |
| SSH | Submenu to open SSH connections in terminal |
| Refresh Ticket | Request/refresh the service ticket for current SPN |
| Copy HTTP Header | Copy `Negotiate <base64-token>` to clipboard |
| Copy Token | Copy raw base64 token to clipboard |
| Debug Mode | Toggle verbose debug output |
| Reload Config | Reload configuration from file |
| About | Shows version, commit, and build date |
| Quit | Exit the application |

## Global Hotkeys

krb5tray supports global hotkeys for quick access to snippets and URLs. Hold the modifier keys and press digits to select by index:

### Snippet Hotkeys

| Platform | Hotkey | Action |
|----------|--------|--------|
| macOS | `Cmd+Option+[digits]` | Copy snippet with matching index |
| Windows | `Ctrl+Alt+[digits]` | Copy snippet with matching index |
| Linux | `Ctrl+Alt+[digits]` | Copy snippet with matching index |

### URL Hotkeys

| Platform | Hotkey | Action |
|----------|--------|--------|
| macOS | `Ctrl+Cmd+[digits]` | Open URL with matching index |
| Windows | `Ctrl+Shift+[digits]` | Open URL with matching index |
| Linux | `Ctrl+Shift+[digits]` | Open URL with matching index |

### SSH Hotkeys

| Platform | Hotkey | Action |
|----------|--------|--------|
| macOS | `Ctrl+Option+[digits]` | Open SSH connection with matching index |
| Windows | `Alt+Shift+[digits]` | Open SSH connection with matching index |
| Linux | `Alt+Shift+[digits]` | Open SSH connection with matching index |

### Multi-digit Support

All hotkeys (snippets, URLs, SSH) support multi-digit input:

- Keep the modifier keys held down
- Press digits in sequence (e.g., `1` then `2` for index 12)
- The action triggers automatically after 1 second of no additional input

**Snippet Examples (macOS):**
- `Cmd+Option+5` → copies snippet with index 5
- `Cmd+Option+1` then `2` (keep modifiers held) → copies snippet with index 12
- `Cmd+Option+1` then `2` then `3` → copies snippet with index 123

**URL Examples (macOS):**
- `Ctrl+Cmd+0` → opens URL with index 0
- `Ctrl+Cmd+1` then `2` (keep modifiers held) → opens URL with index 12

**SSH Examples (macOS):**
- `Ctrl+Option+0` → opens SSH connection with index 0
- `Ctrl+Option+1` then `2` (keep modifiers held) → opens SSH with index 12

All items (snippets, URLs, SSH) use the `index` field from config for hotkey access.

**Note:** On macOS, the terminal running the binary needs Accessibility permissions. Go to System Settings → Privacy & Security → Accessibility and add Terminal.app (or your terminal of choice).

## Example Workflow

1. Start the tray application with your SPN:
   ```bash
   export KRB5_SPN=HTTP/myserver.example.com
   ./krb5tray
   ```

2. Click "Get Ticket" in the tray menu

3. Click "Copy HTTP Header"

4. Use in curl or other HTTP client:
   ```bash
   curl -H "Authorization: $(pbpaste)" https://myserver.example.com/api/endpoint
   ```

## Clipboard Support

| Platform | Method |
|----------|--------|
| macOS | Native NSPasteboard API (no external binaries) |
| Windows | `clip` (built-in) |
| Linux | `xclip` or `xsel` (install separately) |

**Linux clipboard tools:**

```bash
# Ubuntu/Debian
sudo apt-get install xclip
# or
sudo apt-get install xsel

# Fedora/RHEL
sudo dnf install xclip
# or
sudo dnf install xsel
```

## Troubleshooting

### "No ticket" or "Error: unsupported platform"
- Ensure you have a valid Kerberos ticket (`klist` to check)
- On macOS, ensure you're running macOS 11 or later
- On Linux, ensure `KRB5CCNAME` points to a valid ccache file

### "Error: failed to connect"
- macOS: Check that GSSCred service is running
- Windows: Ensure you're logged into a domain or have valid LSA credentials
- Linux: Verify `/etc/krb5.conf` is properly configured

### Linux build fails
- Ensure GTK3 development libraries are installed
- Ensure CGO is enabled: `CGO_ENABLED=1 go build`

### Clipboard not working (Linux)
- Install `xclip` or `xsel`
- Ensure X11 or Wayland clipboard is accessible

## License

MIT
