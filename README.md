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
    {"name": "Jira", "url": "https://jira.example.com"},
    {"name": "Confluence", "url": "https://confluence.example.com"}
  ],
  "snippets": [
    {"index": 1, "name": "Bearer Token", "value": "Bearer abc123..."},
    {"index": 2, "name": "API Key", "value": "x-api-key: secret123"}
  ]
}
```

| Section | Description |
|---------|-------------|
| `spns` | Service Principal Names for Kerberos tickets |
| `secrets` | CSM secret configurations |
| `urls` | URL bookmarks that open in browser |
| `snippets` | Text snippets copied to clipboard (use `index` 0-9 for hotkey access) |

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
| Refresh Ticket | Request/refresh the service ticket for current SPN |
| Copy HTTP Header | Copy `Negotiate <base64-token>` to clipboard |
| Copy Token | Copy raw base64 token to clipboard |
| Debug Mode | Toggle verbose debug output |
| Reload Config | Reload configuration from file |
| About | Shows version, commit, and build date |
| Quit | Exit the application |

## Global Hotkeys

krb5tray supports global hotkeys for quick access to frequently-used snippets. Snippets with an `index` field (0-9) can be copied directly without opening the menu:

| Platform | Hotkey | Action |
|----------|--------|--------|
| macOS | `Cmd+Option+[0-9]` | Copy snippet with matching index |
| Windows | `Ctrl+Alt+[0-9]` | Copy snippet with matching index |
| Linux | `Ctrl+Alt+[0-9]` | Copy snippet with matching index |

**Hybrid approach:**
- **Quick access (hotkeys):** Assign `index` values 0-9 to your most frequently used snippets for instant hotkey access
- **Full list (menu):** Click the tray icon → Snippets to see and select from all configured snippets

**Example config:**
```json
{
  "snippets": [
    {"index": 1, "name": "API Token", "value": "Bearer abc123"},
    {"index": 2, "name": "SSH Key", "value": "ssh-rsa AAAA..."},
    {"name": "Rarely used snippet", "value": "..."}
  ]
}
```
The first two snippets are accessible via `Cmd+Option+1` and `Cmd+Option+2`. The third snippet (no index) is only available through the menu.

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
| macOS | `pbcopy` (built-in) |
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
