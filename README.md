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

krb5tray uses a JSON configuration file located at `~/.config/krb5tray.json`. The file supports the following sections:

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
