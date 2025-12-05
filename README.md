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

### Setting the SPN

Before running, set the target SPN:

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
| SPN: ... | Displays current SPN (click to update from environment) |
| Get Ticket | Request/refresh the service ticket |
| Copy HTTP Header | Copy `Negotiate <base64-token>` to clipboard |
| Copy Token | Copy raw base64 token to clipboard |
| Debug Mode | Toggle verbose debug output |
| Quit | Exit the application |

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
