package gottyclient

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// HostConfig represents configuration for a specific host
type HostConfig struct {
	Host            string
	URL             string
	Callsign        string
	User            string
	Password        string
	AdminPassword   string
	SkipTLSVerify   bool
	UseProxyFromEnv bool
	WSOrigin        string
	V2              bool
	PathSuffix      string
}

// Config represents the entire configuration file
type Config struct {
	Hosts map[string]*HostConfig
}

// GetDefaultConfigPath returns the default config file path
func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".gotty-client", "config")
}

// EnsureConfigExists creates the config file with examples if it doesn't exist
func EnsureConfigExists() error {
	configPath := GetDefaultConfigPath()
	if configPath == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // File already exists
	}

	// Create config file with examples
	exampleConfig := `# GoTTY Client Configuration
# Similar to SSH config, this file allows you to define connection settings
# for different hosts. You can then connect using: uberterm <host-alias>
#
# File location: ~/.gotty-client/config
# Permissions: This file should be readable only by you (chmod 600)

# Example: Local development server
#Host local
#    URL http://localhost:8080
#    User admin
#    Password mypassword
#    AdminPassword secretadmin
#    SkipTLSVerify false
#    UseProxyFromEnv false
#    PathSuffix /terminal/

# Example: Production server with TLS
#Host production
#    URL https://gotty.example.com:8080
#    User operator
#    AdminPassword prodadminpass
#    SkipTLSVerify false
#    WSOrigin https://gotty.example.com:8080
#    PathSuffix /terminal/

# Example: UberSDR instance by callsign
#Host m9psy
#    Callsign M9PSY
#    AdminPassword mypassword
#    PathSuffix /terminal/

# Example: Development server with self-signed cert
#Host dev
#    URL https://dev.example.com:8080
#    AdminPassword devpass
#    SkipTLSVerify true
#    V2 true
#    PathSuffix /terminal/

# Example: Using wildcards for multiple hosts
#Host *.internal
#    User admin
#    SkipTLSVerify true
#    UseProxyFromEnv true

# Default settings for all hosts (lowest priority)
#Host *
#    SkipTLSVerify false
#    UseProxyFromEnv false
#    V2 false
#    PathSuffix /terminal/

# Configuration Options:
#   Host            - Alias name for this configuration
#   URL             - Full URL to the GoTTY server (required unless Callsign is set)
#   Callsign        - UberSDR instance callsign (alternative to URL)
#   User            - Username for basic authentication
#   Password        - Password for basic authentication
#   AdminPassword   - Admin password for X-Admin-Password header
#   SkipTLSVerify   - Skip TLS certificate verification (true/false)
#   UseProxyFromEnv - Use HTTP_PROXY/HTTPS_PROXY from environment (true/false)
#   WSOrigin        - WebSocket Origin URL
#   V2              - Use GoTTY 2.0 protocol (true/false)
#   PathSuffix      - Path to append to URL (default: /terminal/)
`

	if err := os.WriteFile(configPath, []byte(exampleConfig), 0600); err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}

	logrus.Infof("Created config file with examples at: %s", configPath)
	return nil
}

// LoadConfig loads the configuration from the default location
func LoadConfig() (*Config, error) {
	configPath := GetDefaultConfigPath()
	if configPath == "" {
		return &Config{Hosts: make(map[string]*HostConfig)}, nil
	}

	return LoadConfigFromPath(configPath)
}

// LoadConfigFromPath loads configuration from a specific file path
func LoadConfigFromPath(path string) (*Config, error) {
	config := &Config{
		Hosts: make(map[string]*HostConfig),
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return empty config if file doesn't exist
		}
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentHost *HostConfig
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse Host directive
		if strings.HasPrefix(line, "Host ") {
			hostName := strings.TrimSpace(strings.TrimPrefix(line, "Host "))
			if hostName == "" {
				return nil, fmt.Errorf("line %d: Host directive requires a name", lineNum)
			}
			currentHost = &HostConfig{
				Host: hostName,
			}
			config.Hosts[hostName] = currentHost
			continue
		}

		// Parse configuration options
		if currentHost == nil {
			return nil, fmt.Errorf("line %d: configuration option outside of Host block", lineNum)
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: invalid configuration line: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "URL":
			currentHost.URL = value
		case "Callsign":
			currentHost.Callsign = value
		case "User":
			currentHost.User = value
		case "Password":
			currentHost.Password = value
		case "AdminPassword":
			currentHost.AdminPassword = value
		case "SkipTLSVerify":
			currentHost.SkipTLSVerify = parseBool(value)
		case "UseProxyFromEnv":
			currentHost.UseProxyFromEnv = parseBool(value)
		case "WSOrigin":
			currentHost.WSOrigin = value
		case "V2":
			currentHost.V2 = parseBool(value)
		case "PathSuffix":
			currentHost.PathSuffix = value
		default:
			logrus.Warnf("line %d: unknown configuration option: %s", lineNum, key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	return config, nil
}

// GetHostConfig returns the configuration for a specific host
// Supports wildcard matching similar to SSH config
func (c *Config) GetHostConfig(hostAlias string) *HostConfig {
	// First try exact match
	if host, ok := c.Hosts[hostAlias]; ok {
		return host
	}

	// Try wildcard matches
	for pattern, host := range c.Hosts {
		if matchPattern(pattern, hostAlias) {
			return host
		}
	}

	// Try default wildcard
	if host, ok := c.Hosts["*"]; ok {
		return host
	}

	return nil
}

// MergeHostConfigs merges multiple host configs with priority
// Later configs override earlier ones
func MergeHostConfigs(configs ...*HostConfig) *HostConfig {
	result := &HostConfig{}

	for _, config := range configs {
		if config == nil {
			continue
		}

		if config.Host != "" {
			result.Host = config.Host
		}
		if config.URL != "" {
			result.URL = config.URL
		}
		if config.Callsign != "" {
			result.Callsign = config.Callsign
		}
		if config.User != "" {
			result.User = config.User
		}
		if config.Password != "" {
			result.Password = config.Password
		}
		if config.AdminPassword != "" {
			result.AdminPassword = config.AdminPassword
		}
		// For boolean fields, we need to track if they were explicitly set
		// For simplicity, we'll just use the last non-zero value
		result.SkipTLSVerify = result.SkipTLSVerify || config.SkipTLSVerify
		result.UseProxyFromEnv = result.UseProxyFromEnv || config.UseProxyFromEnv
		result.V2 = result.V2 || config.V2
		if config.WSOrigin != "" {
			result.WSOrigin = config.WSOrigin
		}
		if config.PathSuffix != "" {
			result.PathSuffix = config.PathSuffix
		}
	}

	return result
}

// ApplyConfigToClient applies a HostConfig to a Client
func (hc *HostConfig) ApplyToClient(client *Client) {
	if hc == nil {
		return
	}

	// Note: URL and Callsign are handled in createClient() before this is called
	// We don't override URL here to avoid duplicate callsign resolution
	
	if hc.User != "" {
		client.User = hc.User
	}
	if hc.Password != "" {
		client.Password = hc.Password
	}
	if hc.AdminPassword != "" {
		client.AdminPassword = hc.AdminPassword
	}
	if hc.SkipTLSVerify {
		client.SkipTLSVerify = hc.SkipTLSVerify
	}
	if hc.UseProxyFromEnv {
		client.UseProxyFromEnv = hc.UseProxyFromEnv
	}
	if hc.WSOrigin != "" {
		client.WSOrigin = hc.WSOrigin
	}
	if hc.V2 {
		client.V2 = hc.V2
	}
	if hc.PathSuffix != "" {
		client.PathSuffix = hc.PathSuffix
	}
}

// matchPattern matches a pattern against a string (simple wildcard support)
func matchPattern(pattern, str string) bool {
	if pattern == "*" {
		return true
	}

	// Simple wildcard matching
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(str, suffix)
	}

	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, ".*")
		return strings.HasPrefix(str, prefix)
	}

	return pattern == str
}

// parseBool parses a boolean value from string
func parseBool(s string) bool {
	s = strings.ToLower(s)
	b, err := strconv.ParseBool(s)
	if err != nil {
		// Handle yes/no
		return s == "yes" || s == "y"
	}
	return b
}

// SaveHostConfig saves or updates a host configuration in the config file
func SaveHostConfig(hostAlias string, config *HostConfig) error {
	configPath := GetDefaultConfigPath()
	if configPath == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Load existing config
	existingConfig, err := LoadConfigFromPath(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing config: %v", err)
	}
	if existingConfig == nil {
		existingConfig = &Config{Hosts: make(map[string]*HostConfig)}
	}

	// Update or add the host config
	config.Host = hostAlias
	existingConfig.Hosts[hostAlias] = config

	// Write config back to file
	return WriteConfig(configPath, existingConfig)
}

// WriteConfig writes the entire configuration to a file
func WriteConfig(path string, config *Config) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write header
	fmt.Fprintln(writer, "# GoTTY Client Configuration")
	fmt.Fprintln(writer, "# Auto-generated and manually editable")
	fmt.Fprintln(writer, "# File location:", path)
	fmt.Fprintln(writer)

	// Write each host configuration
	for hostAlias, hostConfig := range config.Hosts {
		fmt.Fprintf(writer, "Host %s\n", hostAlias)
		
		if hostConfig.URL != "" {
			fmt.Fprintf(writer, "    URL %s\n", hostConfig.URL)
		}
		if hostConfig.Callsign != "" {
			fmt.Fprintf(writer, "    Callsign %s\n", hostConfig.Callsign)
		}
		if hostConfig.User != "" {
			fmt.Fprintf(writer, "    User %s\n", hostConfig.User)
		}
		if hostConfig.Password != "" {
			fmt.Fprintf(writer, "    Password %s\n", hostConfig.Password)
		}
		if hostConfig.AdminPassword != "" {
			fmt.Fprintf(writer, "    AdminPassword %s\n", hostConfig.AdminPassword)
		}
		if hostConfig.SkipTLSVerify {
			fmt.Fprintf(writer, "    SkipTLSVerify true\n")
		}
		if hostConfig.UseProxyFromEnv {
			fmt.Fprintf(writer, "    UseProxyFromEnv true\n")
		}
		if hostConfig.WSOrigin != "" {
			fmt.Fprintf(writer, "    WSOrigin %s\n", hostConfig.WSOrigin)
		}
		if hostConfig.V2 {
			fmt.Fprintf(writer, "    V2 true\n")
		}
		if hostConfig.PathSuffix != "" {
			fmt.Fprintf(writer, "    PathSuffix %s\n", hostConfig.PathSuffix)
		}
		
		fmt.Fprintln(writer)
	}

	return writer.Flush()
}
