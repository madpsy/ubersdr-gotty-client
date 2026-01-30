# Uberterm - Enhanced GoTTY Client

Uberterm is an enhanced version of gotty-client with full support for ubersdr-gotty's session management features.

## Features

- ✅ **Session Management** - List and destroy tmux sessions
- ✅ **X-Admin-Password Header** - Support for admin authentication
- ✅ **Basic Authentication** - Username/password authentication
- ✅ **TLS Support** - Secure connections with optional certificate verification
- ✅ **Proxy Support** - HTTP/HTTPS proxy from environment
- ✅ **GoTTY v2 Protocol** - Compatible with GoTTY 2.0
- ✅ **Detach Keys** - Configurable key sequences for detaching

## Quick Start

### Build

```bash
./build.sh
```

This creates the `uberterm` binary in the current directory.

### Connect to a Terminal

```bash
./uberterm http://localhost:8080
```

### List Sessions

```bash
./uberterm sessions http://localhost:8080
```

### Destroy a Session

```bash
./uberterm destroy http://localhost:8080 session-name
```

## Installation

### From Source

```bash
git clone <repository-url>
cd ubersdr-gotty-client
./build.sh
sudo cp uberterm /usr/local/bin/
```

### Manual Build

```bash
go build -o uberterm ./cmd/gotty-client
```

## Usage

### Basic Connection

```bash
# Connect to a GoTTY server
uberterm http://localhost:8080

# With TLS
uberterm https://example.com:8080

# Skip TLS verification
uberterm --skip-tls-verify https://example.com:8080
```

### Authentication

```bash
# Basic authentication (prompts for password)
uberterm --user admin http://localhost:8080

# Admin password header
uberterm --admin-password secret http://localhost:8080

# Both authentication methods
uberterm --user admin --admin-password secret http://localhost:8080

# Using environment variables
export GOTTY_CLIENT_USER="admin"
export GOTTY_CLIENT_ADMIN_PASSWORD="secret"
uberterm http://localhost:8080
```

### Session Management

```bash
# List all sessions
uberterm sessions http://localhost:8080

# List sessions with authentication
uberterm --admin-password secret sessions http://localhost:8080

# Destroy a session
uberterm destroy http://localhost:8080 session-name

# Destroy with authentication
uberterm --admin-password secret destroy http://localhost:8080 session-name
```

### Advanced Options

```bash
# Enable debug mode
uberterm --debug http://localhost:8080

# Use proxy from environment
uberterm --use-proxy-from-env http://localhost:8080

# Custom WebSocket origin
uberterm --ws-origin "https://example.com" http://localhost:8080

# Custom detach keys
uberterm --detach-keys "ctrl-a,d" http://localhost:8080

# GoTTY v2 protocol
uberterm --v2 http://localhost:8080
```

## Commands

### `uberterm [OPTIONS] URL`

Connect to a GoTTY terminal server.

**Options:**
- `--debug, -D` - Enable debug logging
- `--skip-tls-verify` - Skip TLS certificate verification
- `--use-proxy-from-env` - Use HTTP/HTTPS proxy from environment
- `--v2` - Use GoTTY 2.0 protocol
- `--ws-origin, -w` - WebSocket Origin URL
- `--user, -u` - Username for Basic Authentication
- `--admin-password, -a` - Admin password for X-Admin-Password header
- `--detach-keys` - Key sequence for detaching (default: "ctrl-p,ctrl-q")

### `uberterm sessions [OPTIONS] URL`

List all available tmux sessions on the server.

**Aliases:** `s`

**Example:**
```bash
uberterm sessions http://localhost:8080
```

**Output:**
```
Found 3 session(s):

NAME                 WINDOW               WINDOWS    ATTACHED   CREATED              LAST ACTIVE         
------------------------------------------------------------------------------------------------------------------------
session1             bash                 2          yes        2026-01-30 19:30:15  2026-01-30 19:35:20
session2             vim                  1          no         2026-01-30 18:45:10  2026-01-30 19:20:05
session3             htop                 3          yes        2026-01-30 17:15:30  2026-01-30 19:40:12
```

### `uberterm destroy [OPTIONS] URL SESSION_NAME`

Destroy (kill) a specific tmux session.

**Aliases:** `d`

**Example:**
```bash
uberterm destroy http://localhost:8080 session1
```

**Output:**
```
✓ Session 'session1' destroyed successfully
```

## Environment Variables

- `GOTTY_CLIENT_DEBUG` - Enable debug mode (set to any value)
- `SKIP_TLS_VERIFY` - Skip TLS verification (set to any value)
- `USE_PROXY_FROM_ENV` - Use proxy from environment (set to any value)
- `GOTTY_CLIENT_GOTTY2` - Use GoTTY 2.0 protocol (set to any value)
- `GOTTY_CLIENT_WS_ORIGIN` - WebSocket Origin URL
- `GOTTY_CLIENT_USER` - Username for Basic Authentication
- `GOTTY_CLIENT_ADMIN_PASSWORD` - Admin password for X-Admin-Password header

## Integration with ubersdr-gotty

Uberterm is specifically designed to work with [ubersdr-gotty](https://github.com/yourusername/ubersdr-gotty), which provides:

- **Persistent Sessions** - tmux sessions that survive disconnections
- **Session API** - REST API for listing and managing sessions
- **Admin Authentication** - X-Admin-Password header support
- **Web Interface** - Browser-based session management

### Workflow Example

1. **Start ubersdr-gotty server:**
   ```bash
   cd /home/nathan/repos/ubersdr-gotty
   ./run.sh
   ```

2. **List available sessions:**
   ```bash
   uberterm sessions http://localhost:8080
   ```

3. **Connect to a terminal:**
   ```bash
   uberterm http://localhost:8080
   ```

4. **Work in the terminal, then detach:**
   - Press `Ctrl+P`, then `Ctrl+Q` (default detach keys)
   - Or close the terminal window

5. **Reconnect later:**
   - Use the web interface at `http://localhost:8080/sessions`
   - Or connect via uberterm again

6. **Clean up old sessions:**
   ```bash
   uberterm destroy http://localhost:8080 old-session
   ```

## Detaching from a Session

By default, you can detach from a terminal session using `Ctrl+P` followed by `Ctrl+Q`. This leaves the session running on the server.

To customize the detach keys:
```bash
uberterm --detach-keys "ctrl-a,d" http://localhost:8080
```

## Troubleshooting

### Connection Issues

```bash
# Enable debug logging
uberterm --debug http://localhost:8080

# Skip TLS verification for self-signed certificates
uberterm --skip-tls-verify https://localhost:8080
```

### Authentication Failures

```bash
# Verify credentials
uberterm --user admin http://localhost:8080
# (will prompt for password)

# Check admin password
uberterm --admin-password secret sessions http://localhost:8080
```

### Session Not Found

```bash
# List all sessions first
uberterm sessions http://localhost:8080

# Then destroy with exact name
uberterm destroy http://localhost:8080 exact-session-name
```

## Development

### Project Structure

```
ubersdr-gotty-client/
├── cmd/gotty-client/     # CLI application
│   └── main.go           # Command-line interface
├── gotty-client.go       # Core client library
├── build.sh              # Build script
├── SESSIONS.md           # Session features documentation
└── README_UBERTERM.md    # This file
```

### Building from Source

```bash
# Install dependencies
go mod download

# Build
go build -o uberterm ./cmd/gotty-client

# Or use the build script
./build.sh
```

### Running Tests

```bash
go test ./...
```

## API Reference

See [SESSIONS.md](SESSIONS.md) for detailed API documentation.

## License

See [LICENSE](LICENSE) file for details.

## Credits

Based on [gotty-client](https://github.com/moul/gotty-client) by Manfred Touron.

Enhanced with session management features for [ubersdr-gotty](https://github.com/yourusername/ubersdr-gotty).
