# Uberterm - Session Management Features

This document describes the session management features in uberterm (gotty-client) to support ubersdr-gotty's session functionality.

## New Features

### 1. X-Admin-Password Header Support

The client now supports passing an admin password via the `X-Admin-Password` header for authentication with the GoTTY server.

**Usage:**
```bash
# Via command line flag
uberterm --admin-password "your-password" http://localhost:8080

# Via environment variable
export GOTTY_CLIENT_ADMIN_PASSWORD="your-password"
uberterm http://localhost:8080
```

### 2. Session Listing

List all available tmux sessions on the GoTTY server.

**Usage:**
```bash
# List sessions
uberterm sessions http://localhost:8080

# With admin password
uberterm --admin-password "your-password" sessions http://localhost:8080

# With basic auth
uberterm --user admin --admin-password "your-password" sessions http://localhost:8080
```

**Example Output:**
```
Found 3 session(s):

NAME                 WINDOW               WINDOWS    ATTACHED   CREATED              LAST ACTIVE         
------------------------------------------------------------------------------------------------------------------------
session1             bash                 2          yes        2026-01-30 19:30:15  2026-01-30 19:35:20
session2             vim                  1          no         2026-01-30 18:45:10  2026-01-30 19:20:05
session3             htop                 3          yes        2026-01-30 17:15:30  2026-01-30 19:40:12
```

### 3. Session Destruction

Destroy (kill) a specific tmux session by name.

**Usage:**
```bash
# Destroy a session
uberterm destroy http://localhost:8080 session1

# With admin password
uberterm --admin-password "your-password" destroy http://localhost:8080 session1

# With basic auth
uberterm --user admin --admin-password "your-password" destroy http://localhost:8080 session1
```

**Example Output:**
```
✓ Session 'session1' destroyed successfully
```

## API Endpoints

The client now interacts with the following API endpoints on the GoTTY server:

- `GET /api/sessions` - List all tmux sessions
- `DELETE /api/sessions/destroy?name=<session_name>` - Destroy a specific session

## Authentication

The client supports multiple authentication methods:

1. **Basic Authentication** - Using `--user` flag (prompts for password)
2. **Admin Password Header** - Using `--admin-password` flag or `GOTTY_CLIENT_ADMIN_PASSWORD` environment variable
3. **Combined** - Both basic auth and admin password can be used together

## Command Reference

### Main Command (Terminal Connection)
```bash
uberterm [OPTIONS] GOTTY_URL
```

### Sessions Command
```bash
uberterm sessions [OPTIONS] GOTTY_URL
uberterm s [OPTIONS] GOTTY_URL  # Short alias
```

### Destroy Command
```bash
uberterm destroy [OPTIONS] GOTTY_URL SESSION_NAME
uberterm d [OPTIONS] GOTTY_URL SESSION_NAME  # Short alias
```

## Global Options

- `--debug, -D` - Enable debug mode
- `--skip-tls-verify` - Skip TLS certificate verification
- `--use-proxy-from-env` - Use proxy settings from environment
- `--v2` - Use GoTTY 2.0 protocol
- `--ws-origin, -w` - WebSocket Origin URL
- `--user, -u` - Username for Basic Authentication
- `--admin-password, -a` - Admin password for X-Admin-Password header
- `--detach-keys` - Key sequence for detaching (default: "ctrl-p,ctrl-q")

## Environment Variables

- `GOTTY_CLIENT_DEBUG` - Enable debug mode
- `SKIP_TLS_VERIFY` - Skip TLS verification
- `USE_PROXY_FROM_ENV` - Use proxy from environment
- `GOTTY_CLIENT_GOTTY2` - Use GoTTY 2.0 protocol
- `GOTTY_CLIENT_WS_ORIGIN` - WebSocket Origin URL
- `GOTTY_CLIENT_USER` - Username for Basic Authentication
- `GOTTY_CLIENT_ADMIN_PASSWORD` - Admin password for X-Admin-Password header

## Examples

### Connect to a normal terminal
```bash
uberterm http://localhost:8080
```

### Connect with admin password
```bash
uberterm --admin-password "secret" http://localhost:8080
```

### List sessions with authentication
```bash
uberterm --user admin --admin-password "secret" sessions http://localhost:8080
```

### Destroy a session
```bash
uberterm --admin-password "secret" destroy http://localhost:8080 old-session
```

### Using environment variables
```bash
export GOTTY_CLIENT_ADMIN_PASSWORD="secret"
export GOTTY_CLIENT_USER="admin"

uberterm sessions http://localhost:8080
uberterm destroy http://localhost:8080 old-session
uberterm http://localhost:8080
```

## Integration with ubersdr-gotty

This client is designed to work with ubersdr-gotty's session management features:

1. **Session Persistence** - Sessions created on the server persist even after disconnection
2. **Session Reconnection** - Connect to existing sessions by navigating to the sessions page
3. **Session Management** - List and destroy sessions via the API
4. **Admin Authentication** - Use X-Admin-Password header for administrative operations

## Building

```bash
cd /home/nathan/repos/ubersdr-gotty-client
./build.sh
```

This will create the `uberterm` binary in the current directory.

## Testing

```bash
# Start ubersdr-gotty server
cd /home/nathan/repos/ubersdr-gotty
./run.sh

# In another terminal, test the client
cd /home/nathan/repos/ubersdr-gotty-client

# Build the client
./build.sh

# List sessions
./uberterm sessions http://localhost:8080

# Connect to terminal
./uberterm http://localhost:8080

# Destroy a session
./uberterm destroy http://localhost:8080 session-name
```

## Quick Start

1. **Build the client:**
   ```bash
   cd /home/nathan/repos/ubersdr-gotty-client
   ./build.sh
   ```

2. **Connect to a GoTTY server:**
   ```bash
   ./uberterm http://localhost:8080
   ```

3. **List available sessions:**
   ```bash
   ./uberterm sessions http://localhost:8080
   ```

4. **Clean up old sessions:**
   ```bash
   ./uberterm destroy http://localhost:8080 old-session-name
   ```
