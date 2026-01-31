# Configuration File Guide

Uberterm supports SSH-style configuration files for managing connection settings, including admin passwords and other credentials.

## Quick Start

On first run, uberterm automatically creates a config file at `~/.gotty-client/config` with commented examples.

To use a host alias:
```bash
uberterm myserver
```

## Configuration File Location

Default: `~/.gotty-client/config`

Custom location:
```bash
uberterm --config /path/to/config myserver
```

## File Format

The configuration file uses an SSH-style format with `Host` blocks:

```
Host <alias>
    <option> <value>
    <option> <value>
```

## Configuration Options

| Option | Description | Example |
|--------|-------------|---------|
| `URL` | Full URL to the GoTTY server (required) | `http://localhost:8080` |
| `User` | Username for basic authentication | `admin` |
| `Password` | Password for basic authentication | `mypassword` |
| `AdminPassword` | Admin password for X-Admin-Password header | `secretadmin` |
| `SkipTLSVerify` | Skip TLS certificate verification | `true` or `false` |
| `UseProxyFromEnv` | Use HTTP_PROXY/HTTPS_PROXY from environment | `true` or `false` |
| `WSOrigin` | WebSocket Origin URL | `http://localhost:8080` |
| `V2` | Use GoTTY 2.0 protocol | `true` or `false` |

## Example Configuration

### Basic Local Server

```
Host local
    URL http://localhost:8080
    AdminPassword secretadmin
```

Usage:
```bash
uberterm local
uberterm sessions local
uberterm destroy local session-name
```

### Production Server with Authentication

```
Host production
    URL https://gotty.example.com:8080
    User operator
    Password prodpass
    AdminPassword prodadminpass
    SkipTLSVerify false
```

Usage:
```bash
uberterm production
```

### Development Server with Self-Signed Certificate

```
Host dev
    URL https://dev.example.com:8080
    AdminPassword devpass
    SkipTLSVerify true
    V2 true
```

### Multiple Servers with Wildcards

```
# All internal servers
Host *.internal
    User admin
    SkipTLSVerify true
    UseProxyFromEnv true

# Specific internal server
Host server1.internal
    URL https://server1.internal:8080
    AdminPassword server1pass

# Another internal server
Host server2.internal
    URL https://server2.internal:8080
    AdminPassword server2pass
```

Usage:
```bash
uberterm server1.internal
uberterm server2.internal
```

### Default Settings for All Hosts

```
Host *
    SkipTLSVerify false
    UseProxyFromEnv false
    V2 false
```

## Priority Order

Settings are applied in the following order (later overrides earlier):

1. **Default wildcard** (`Host *`) - Lowest priority
2. **Wildcard matches** (e.g., `Host *.internal`)
3. **Exact host match** (e.g., `Host production`)
4. **Command-line flags** - Highest priority

Example:
```
Host *
    SkipTLSVerify false

Host production
    URL https://prod.example.com:8080
    SkipTLSVerify true
```

```bash
# Uses SkipTLSVerify=true from config
uberterm production

# Overrides to SkipTLSVerify=false via command-line
uberterm --skip-tls-verify=false production
```

## Security Considerations

### File Permissions

The config file is automatically created with `0600` permissions (readable/writable only by owner):

```bash
chmod 600 ~/.gotty-client/config
```

### Storing Passwords

While the config file supports storing passwords, consider these alternatives:

1. **Omit passwords** - You'll be prompted when needed:
   ```
   Host production
       URL https://prod.example.com:8080
       User operator
       # Password omitted - will prompt
   ```

2. **Use environment variables**:
   ```bash
   export GOTTY_CLIENT_ADMIN_PASSWORD="secret"
   uberterm production
   ```

3. **Use command-line flags** (not recommended for scripts):
   ```bash
   uberterm --admin-password secret production
   ```

### Best Practices

- Keep config file permissions at `0600`
- Don't commit config files with passwords to version control
- Use `.gitignore` to exclude config files
- Consider using a password manager or secrets vault
- Rotate passwords regularly

## Using with Scripts

### Connect to a specific host

```bash
#!/bin/bash
uberterm production
```

### List sessions on multiple hosts

```bash
#!/bin/bash
for host in server1 server2 server3; do
    echo "Sessions on $host:"
    uberterm sessions $host
    echo
done
```

### Automated session cleanup

```bash
#!/bin/bash
HOST="production"
OLD_SESSIONS=$(uberterm sessions $HOST | grep "no" | awk '{print $1}')

for session in $OLD_SESSIONS; do
    echo "Destroying detached session: $session"
    uberterm destroy $HOST $session
done
```

## Troubleshooting

### Config file not found

If the config file isn't automatically created:

```bash
mkdir -p ~/.gotty-client
touch ~/.gotty-client/config
chmod 600 ~/.gotty-client/config
```

### Host alias not found

```
Error: host alias 'myserver' not found in config file
```

Check that:
1. The host is defined in your config file
2. The `Host` directive matches exactly
3. The config file path is correct

### URL not configured

```
Error: host alias 'myserver' has no URL configured
```

Ensure the host block has a `URL` option:
```
Host myserver
    URL http://localhost:8080
```

### Permission denied

```
Error: failed to open config file: permission denied
```

Fix permissions:
```bash
chmod 600 ~/.gotty-client/config
```

## Migration from Command-Line Usage

### Before (command-line only)

```bash
uberterm --user admin --admin-password secret --skip-tls-verify https://server.example.com:8080
```

### After (using config)

Create config entry:
```
Host myserver
    URL https://server.example.com:8080
    User admin
    AdminPassword secret
    SkipTLSVerify true
```

Use alias:
```bash
uberterm myserver
```

## Advanced Examples

### Multiple Environments

```
# Development
Host dev
    URL http://localhost:8080
    AdminPassword devpass

# Staging
Host staging
    URL https://staging.example.com:8080
    User operator
    AdminPassword stagingpass
    SkipTLSVerify true

# Production
Host prod
    URL https://prod.example.com:8080
    User operator
    AdminPassword prodpass
    SkipTLSVerify false
```

### Team Configuration

```
# Shared team server
Host team
    URL https://team.example.com:8080
    User teamuser
    # Password will be prompted

# Personal development
Host mydev
    URL http://localhost:8080
    AdminPassword mydevpass
```

### Proxy Configuration

```
Host corporate
    URL https://gotty.corp.example.com:8080
    User employee
    UseProxyFromEnv true
    SkipTLSVerify true
```

Then set proxy environment variables:
```bash
export HTTP_PROXY=http://proxy.corp.example.com:8080
export HTTPS_PROXY=http://proxy.corp.example.com:8080
uberterm corporate
```

## See Also

- [README_UBERTERM.md](README_UBERTERM.md) - Main documentation
- [SESSIONS.md](SESSIONS.md) - Session management API
- SSH config documentation for similar syntax
