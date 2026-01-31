package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"syscall"
	"time"

	gottyclient "github.com/moul/gotty-client"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

var VERSION = "dev"

func main() {
	app := cli.NewApp()
	app.Name = "uberterm"
	app.Usage = "GoTTY client for your terminal with session management"
	app.Version = VERSION
	app.Author = "Enhanced for ubersdr-gotty"

	// Ensure config file exists on startup
	if err := gottyclient.EnsureConfigExists(); err != nil {
		logrus.Warnf("Failed to ensure config file exists: %v", err)
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug, D",
			Usage:  "Enable debug mode",
			EnvVar: "GOTTY_CLIENT_DEBUG",
		},
		cli.BoolFlag{
			Name:   "skip-tls-verify",
			Usage:  "Skip TLS verify",
			EnvVar: "SKIP_TLS_VERIFY",
		},
		cli.BoolFlag{
			Name:   "use-proxy-from-env",
			Usage:  "Use Proxy from environment",
			EnvVar: "USE_PROXY_FROM_ENV",
		},
		cli.StringFlag{
			Name:  "detach-keys",
			Value: "ctrl-p,ctrl-q",
			Usage: "Key sequence for detaching gotty-client",
		},
		cli.BoolFlag{
			Name:   "v2",
			Usage:  "For Gotty 2.0",
			EnvVar: "GOTTY_CLIENT_GOTTY2",
		},
		cli.StringFlag{
			Name:   "ws-origin, w",
			Usage:  "WebSocket Origin URL",
			EnvVar: "GOTTY_CLIENT_WS_ORIGIN",
		},
		cli.StringFlag{
			Name:   "user, u",
			Usage:  "Username for Basic Authentication",
			EnvVar: "GOTTY_CLIENT_USER",
		},
		cli.StringFlag{
			Name:   "password, p",
			Usage:  "Password for Basic Authentication (will prompt if not provided)",
			EnvVar: "GOTTY_CLIENT_PASSWORD",
		},
		cli.StringFlag{
			Name:   "admin-password, a",
			Usage:  "Admin password for X-Admin-Password header",
			EnvVar: "GOTTY_CLIENT_ADMIN_PASSWORD",
		},
		cli.StringFlag{
			Name:   "session, s",
			Usage:  "Tmux session name to attach to",
			EnvVar: "GOTTY_CLIENT_SESSION",
		},
		cli.StringFlag{
			Name:   "window, n",
			Usage:  "Window name to look up session by (auto-resolves to session name)",
			EnvVar: "GOTTY_CLIENT_WINDOW",
		},
		cli.BoolFlag{
			Name:   "new-session",
			Usage:  "Create a new session with auto-generated window name (or use next arg as name)",
			EnvVar: "GOTTY_CLIENT_NEW_SESSION",
		},
		cli.StringFlag{
			Name:   "callsign",
			Usage:  "UberSDR instance callsign to connect to (auto-resolves to URL)",
			EnvVar: "GOTTY_CLIENT_CALLSIGN",
		},
		cli.StringFlag{
			Name:   "path-suffix",
			Usage:  "Path suffix to append to URL (default: /terminal/)",
			Value:  "/terminal/",
			EnvVar: "GOTTY_CLIENT_PATH_SUFFIX",
		},
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Path to config file",
			Value: gottyclient.GetDefaultConfigPath(),
		},
		cli.StringFlag{
			Name:  "save",
			Usage: "Save connection settings to config file with this alias",
		},
		cli.BoolFlag{
			Name:  "list-sessions, ls",
			Usage: "List available tmux sessions",
		},
		cli.BoolFlag{
			Name:  "list-instances, li",
			Usage: "List available UberSDR instances",
		},
		cli.StringFlag{
			Name:  "destroy-session",
			Usage: "Destroy a tmux session by name",
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}

	app.Action = mainAction

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createClient(c *cli.Context) (*gottyclient.Client, error) {
	// Load config file
	config, err := gottyclient.LoadConfigFromPath(c.String("config"))
	if err != nil {
		logrus.Warnf("Failed to load config file: %v", err)
		config = &gottyclient.Config{Hosts: make(map[string]*gottyclient.HostConfig)}
	}

	// Check if callsign is specified
	callsign := ""
	if c.IsSet("callsign") {
		callsign = c.String("callsign")
	} else if c.GlobalIsSet("callsign") {
		callsign = c.GlobalString("callsign")
	}

	var urlOrAlias string
	
	if callsign != "" {
		// Look up instance by callsign
		logrus.Infof("Looking up instance by callsign: %s", callsign)
		instance, err := gottyclient.FindInstanceByCallsign(callsign)
		if err != nil {
			return nil, fmt.Errorf("failed to find instance: %v", err)
		}
		urlOrAlias = instance.PublicURL
		logrus.Infof("Found instance '%s' at %s", instance.Callsign, instance.PublicURL)
	} else {
		// Get URL from arguments
		args := c.Args()
		if len(args) == 0 {
			return nil, fmt.Errorf("URL, host alias, or --callsign required")
		}
		
		// Check if --new-session is set and first arg might be the window name
		if c.Bool("new-session") && len(args) >= 2 {
			// First arg could be window name, second is URL/alias
			// Check if second arg looks like a URL or known alias
			config, _ := gottyclient.LoadConfigFromPath(c.String("config"))
			if config != nil && config.GetHostConfig(args[1]) != nil {
				// Second arg is a known alias, so first arg is window name
				urlOrAlias = args[1]
			} else if strings.HasPrefix(args[1], "http://") || strings.HasPrefix(args[1], "https://") || strings.Contains(args[1], ":") {
				// Second arg looks like a URL, so first arg is window name
				urlOrAlias = args[1]
			} else {
				// Can't determine, use first arg as URL
				urlOrAlias = args[0]
			}
		} else {
			urlOrAlias = args[0]
		}
	}
	
	// Try to get host config from config file
	var hostConfig *gottyclient.HostConfig
	var url string
	
	// Check if it's a URL or a host alias
	if strings.HasPrefix(urlOrAlias, "http://") || strings.HasPrefix(urlOrAlias, "https://") {
		// It's a URL
		url = urlOrAlias
		// Try to get default config
		hostConfig = config.GetHostConfig("*")
	} else {
		// Try to find it as a host alias
		hostConfig = config.GetHostConfig(urlOrAlias)
		if hostConfig != nil {
			// Found a host config - check if it has URL or Callsign
			if hostConfig.URL != "" {
				url = hostConfig.URL
			} else if hostConfig.Callsign != "" {
				// Resolve callsign to URL
				logrus.Infof("Resolving callsign from config: %s", hostConfig.Callsign)
				instance, err := gottyclient.FindInstanceByCallsign(hostConfig.Callsign)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve callsign %s: %v", hostConfig.Callsign, err)
				}
				url = instance.PublicURL
				logrus.Infof("Resolved callsign %s to %s", hostConfig.Callsign, instance.PublicURL)
			} else {
				return nil, fmt.Errorf("host config '%s' has neither URL nor Callsign", urlOrAlias)
			}
		} else {
			// Not a host alias, treat as URL without scheme (backward compatibility)
			// This allows commands like: uberterm localhost:8080
			url = urlOrAlias
			if !strings.Contains(url, "://") {
				url = "http://" + url
			}
			// Try to get default config
			hostConfig = config.GetHostConfig("*")
		}
	}
	
	// Apply path suffix (default: /terminal/)
	pathSuffix := "/terminal/"
	if c.IsSet("path-suffix") {
		pathSuffix = c.String("path-suffix")
	} else if c.GlobalIsSet("path-suffix") {
		pathSuffix = c.GlobalString("path-suffix")
	} else if hostConfig != nil && hostConfig.PathSuffix != "" {
		pathSuffix = hostConfig.PathSuffix
	}
	
	// Append path suffix to URL if it doesn't already have it
	if pathSuffix != "" && !strings.Contains(url, pathSuffix) {
		// Parse URL to add path suffix
		parsedURL, err := gottyclient.ParseURL(url)
		if err == nil {
			// Check if URL already ends with a path
			if !strings.HasSuffix(parsedURL, "/") && !strings.HasPrefix(pathSuffix, "/") {
				parsedURL += "/"
			}
			url = strings.TrimSuffix(parsedURL, "/") + pathSuffix
			logrus.Debugf("Applied path suffix: %s", pathSuffix)
		}
	}
	
	// Check if creating a new session
	createNewSession := c.Bool("new-session") || c.GlobalBool("new-session")
	newSessionName := ""
	
	if createNewSession {
		// Check if window name was provided as first argument
		args := c.Args()
		if len(args) >= 2 {
			// Check if first arg is the window name (second arg is URL/alias)
			config, _ := gottyclient.LoadConfigFromPath(c.String("config"))
			secondArgIsHost := false
			
			if config != nil && config.GetHostConfig(args[1]) != nil {
				secondArgIsHost = true
			} else if strings.HasPrefix(args[1], "http://") || strings.HasPrefix(args[1], "https://") || strings.Contains(args[1], ":") {
				secondArgIsHost = true
			}
			
			if secondArgIsHost {
				newSessionName = args[0]
				logrus.Debugf("Using custom window name from argument: %s", newSessionName)
			}
		}
		
		// If no custom name provided, generate one
		if newSessionName == "" {
			newSessionName = generateSessionName()
			logrus.Infof("Auto-generated window name: %s", newSessionName)
		}
	}
	
	// Determine session name - either from --session or by looking up --window
	sessionName := ""
	if c.IsSet("session") {
		sessionName = c.String("session")
	} else if c.GlobalIsSet("session") {
		sessionName = c.GlobalString("session")
	}
	
	// If --new-session is specified, auto-generate a session ID
	if newSessionName != "" && sessionName == "" {
		// Generate unique session ID using timestamp
		sessionName = fmt.Sprintf("%d", time.Now().Unix())
		logrus.Infof("Auto-generated session ID: %s", sessionName)
	}
	
	// If window name is specified, look up the session
	windowName := ""
	if c.IsSet("window") {
		windowName = c.String("window")
	} else if c.GlobalIsSet("window") {
		windowName = c.GlobalString("window")
	}
	
	if windowName != "" && sessionName == "" && newSessionName == "" {
		// Need to look up session by window name
		logrus.Debugf("Looking up session by window name: %s", windowName)
		
		// Create a temporary client to query sessions
		tempClient, err := gottyclient.NewClient(url)
		if err == nil {
			// Apply authentication settings
			if hostConfig != nil {
				hostConfig.ApplyToClient(tempClient)
			}
			if c.IsSet("admin-password") || c.GlobalIsSet("admin-password") {
				if c.IsSet("admin-password") {
					tempClient.AdminPassword = c.String("admin-password")
				} else {
					tempClient.AdminPassword = c.GlobalString("admin-password")
				}
			}
			if c.IsSet("user") {
				tempClient.User = c.String("user")
			}
			if c.IsSet("password") {
				tempClient.Password = c.String("password")
			}
			
			// Query sessions
			sessions, err := tempClient.ListSessions()
			if err != nil {
				logrus.Warnf("Failed to look up session by window name: %v", err)
			} else {
				// Find session with matching window name
				for _, session := range sessions.Sessions {
					if session.WindowName == windowName {
						sessionName = session.Name
						logrus.Infof("Found session '%s' with window name '%s'", sessionName, windowName)
						break
					}
				}
				if sessionName == "" {
					logrus.Warnf("No session found with window name '%s'", windowName)
				}
			}
		}
	}
	
	// Add session parameter if specified
	if sessionName != "" {
		parsedURL, err := gottyclient.ParseURL(url)
		if err == nil {
			separator := "?"
			if strings.Contains(parsedURL, "?") {
				separator = "&"
			}
			url = parsedURL + separator + "session=" + sessionName
			
			// If this is a new session with custom window name, add name parameter too
			if newSessionName != "" {
				url = url + "&name=" + newSessionName
				logrus.Infof("Creating new session '%s' with window name: %s", sessionName, newSessionName)
				fmt.Println("\n💡 Tip: To detach from session without closing it, press Ctrl-b then d\n")
			} else {
				logrus.Debugf("Attaching to session: %s", sessionName)
				fmt.Println("\n💡 Tip: To detach from session without closing it, press Ctrl-b then d\n")
			}
		}
	}

	// Create client
	client, err := gottyclient.NewClient(url)
	if err != nil {
		return nil, err
	}

	// Apply config file settings (lowest priority)
	if hostConfig != nil {
		hostConfig.ApplyToClient(client)
	}

	// Default to V2 protocol (ubersdr-gotty uses V2)
	client.V2 = true

	// Apply command-line flags (highest priority)
	if c.IsSet("skip-tls-verify") || c.Bool("skip-tls-verify") {
		client.SkipTLSVerify = c.Bool("skip-tls-verify")
	}
	if c.IsSet("use-proxy-from-env") || c.Bool("use-proxy-from-env") {
		client.UseProxyFromEnv = c.Bool("use-proxy-from-env")
	}
	// Allow explicit override of V2 setting
	if c.IsSet("v2") {
		client.V2 = c.Bool("v2")
	}
	if c.IsSet("ws-origin") {
		client.WSOrigin = c.String("ws-origin")
	}
	if c.IsSet("user") {
		client.User = c.String("user")
	}
	if c.IsSet("password") {
		client.Password = c.String("password")
	}
	// Check both flag name and alias for admin-password
	if c.IsSet("admin-password") || c.IsSet("a") {
		client.AdminPassword = c.String("admin-password")
	}
	// Also check global context for subcommands
	if client.AdminPassword == "" && c.GlobalIsSet("admin-password") {
		client.AdminPassword = c.GlobalString("admin-password")
	}
	
	// Set path suffix
	if c.IsSet("path-suffix") {
		client.PathSuffix = c.String("path-suffix")
	} else if c.GlobalIsSet("path-suffix") {
		client.PathSuffix = c.GlobalString("path-suffix")
	}
	
	logrus.Debugf("Client configuration: User=%q, AdminPassword set=%v, PathSuffix=%q", client.User, client.AdminPassword != "", client.PathSuffix)

	// If user is set but password is not, prompt for password
	if client.User != "" && client.Password == "" && !c.IsSet("password") {
		fmt.Printf("Password for %s: ", client.User)
		passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %v", err)
		}
		client.Password = string(passwordBytes)
	}

	// Parse detach keys
	detachKeys := c.String("detach-keys")
	client.EscapeKeys = parseDetachKeys(detachKeys)

	return client, nil
}

func mainAction(c *cli.Context) error {
	// Handle list instances flag
	if c.Bool("list-instances") {
		return listInstancesAction(c)
	}

	// Handle destroy session flag
	if c.IsSet("destroy-session") {
		return destroySessionAction(c)
	}

	// Handle list sessions flag
	if c.Bool("list-sessions") {
		return listSessionsAction(c)
	}

	// Default: connect to terminal
	return connectAction(c)
}

func connectAction(c *cli.Context) error {
	client, err := createClient(c)
	if err != nil {
		return err
	}

	// Save config if --save flag is provided
	if c.IsSet("save") {
		saveAlias := c.String("save")
		if saveAlias == "" {
			return fmt.Errorf("--save requires an alias name")
		}
		
		if err := saveConnectionConfig(c, client, saveAlias); err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}
		
		fmt.Printf("✓ Saved connection settings as '%s' in %s\n", saveAlias, c.String("config"))
	}

	if err := client.Loop(); err != nil {
		return err
	}

	return nil
}

func saveConnectionConfig(c *cli.Context, client *gottyclient.Client, alias string) error {
	// Build host config from current settings
	hostConfig := &gottyclient.HostConfig{
		Host: alias,
	}
	
	// Determine what to save based on how the connection was made
	if c.IsSet("callsign") {
		// Save callsign instead of URL (uppercase for consistency)
		hostConfig.Callsign = strings.ToUpper(c.String("callsign"))
	} else {
		// Save the URL
		hostConfig.URL = client.URL
	}
	
	// Save authentication settings
	if client.User != "" {
		hostConfig.User = client.User
	}
	if client.Password != "" && c.IsSet("password") {
		// Only save password if explicitly provided via flag (not prompted)
		hostConfig.Password = client.Password
	}
	if client.AdminPassword != "" {
		hostConfig.AdminPassword = client.AdminPassword
	}
	
	// Save connection settings
	if client.SkipTLSVerify {
		hostConfig.SkipTLSVerify = true
	}
	if client.UseProxyFromEnv {
		hostConfig.UseProxyFromEnv = true
	}
	if client.WSOrigin != "" {
		hostConfig.WSOrigin = client.WSOrigin
	}
	if client.V2 {
		hostConfig.V2 = true
	}
	if client.PathSuffix != "" {
		hostConfig.PathSuffix = client.PathSuffix
	}
	
	// Save to config file
	return gottyclient.SaveHostConfig(alias, hostConfig)
}

func listInstancesAction(c *cli.Context) error {
	instances, err := gottyclient.ListInstances()
	if err != nil {
		return fmt.Errorf("failed to list instances: %v", err)
	}

	if instances.Count == 0 {
		fmt.Println("No instances found.")
		return nil
	}

	fmt.Printf("Found %d UberSDR instance(s):\n\n", instances.Count)
	fmt.Printf("%-15s %-40s %-30s %-8s %-8s %s\n",
		"CALLSIGN", "NAME", "LOCATION", "CLIENTS", "LOAD", "URL")
	fmt.Println(strings.Repeat("-", 150))

	for _, instance := range instances.Instances {
		fmt.Printf("%-15s %-40s %-30s %-8s %-8s %s\n",
			instance.Callsign,
			truncate(instance.Name, 40),
			truncate(instance.Location, 30),
			fmt.Sprintf("%d/%d", instance.AvailableClients, instance.MaxClients),
			instance.LoadStatus,
			instance.PublicURL)
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func listSessionsAction(c *cli.Context) error {
	client, err := createClient(c)
	if err != nil {
		return err
	}

	sessions, err := client.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %v", err)
	}

	if sessions.Count == 0 {
		fmt.Println("No sessions found.")
		return nil
	}

	fmt.Printf("Found %d session(s):\n\n", sessions.Count)
	fmt.Printf("%-30s %-20s %-8s %-10s %-20s %-20s\n",
		"NAME", "WINDOW", "WINDOWS", "ATTACHED", "CREATED", "LAST ACTIVE")
	fmt.Println(strings.Repeat("-", 130))

	for _, session := range sessions.Sessions {
		attached := "no"
		if session.Attached {
			attached = "yes"
		}
		fmt.Printf("%-30s %-20s %-8d %-10s %-20s %-20s\n",
			session.Name,
			session.WindowName,
			session.Windows,
			attached,
			session.Created,
			session.LastActive)
	}

	return nil
}

func destroySessionAction(c *cli.Context) error {
	sessionName := c.String("destroy-session")
	if sessionName == "" {
		return fmt.Errorf("session name required for --destroy-session")
	}

	client, err := createClient(c)
	if err != nil {
		return err
	}

	resp, err := client.DestroySession(sessionName)
	if err != nil {
		return fmt.Errorf("failed to destroy session: %v", err)
	}

	if resp.Success {
		fmt.Printf("✓ Session '%s' destroyed successfully\n", sessionName)
	} else {
		return fmt.Errorf("failed to destroy session: %s", resp.Message)
	}

	return nil
}

func parseDetachKeys(keys string) []byte {
	parts := strings.Split(keys, ",")
	result := make([]byte, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		switch part {
		case "ctrl-a":
			result = append(result, 1)
		case "ctrl-b":
			result = append(result, 2)
		case "ctrl-c":
			result = append(result, 3)
		case "ctrl-d":
			result = append(result, 4)
		case "ctrl-e":
			result = append(result, 5)
		case "ctrl-p":
			result = append(result, 16)
		case "ctrl-q":
			result = append(result, 17)
		case "ctrl-x":
			result = append(result, 24)
		case "ctrl-z":
			result = append(result, 26)
		default:
			if len(part) == 1 {
				result = append(result, part[0])
			}
		}
	}

	return result
}

// generateSessionName generates a random adjective-noun combination
func generateSessionName() string {
	adjectives := []string{
		"happy", "clever", "brave", "swift", "bright", "calm", "eager", "gentle",
		"jolly", "kind", "lively", "merry", "nice", "proud", "silly", "witty",
		"zany", "bold", "cool", "daring", "fancy", "grand", "lucky", "mighty",
		"noble", "quick", "smart", "wise", "agile", "cosmic", "dynamic", "epic",
	}
	
	nouns := []string{
		"panda", "tiger", "eagle", "dolphin", "falcon", "lion", "wolf", "bear",
		"hawk", "fox", "owl", "shark", "dragon", "phoenix", "unicorn", "griffin",
		"raven", "cobra", "jaguar", "lynx", "otter", "badger", "ferret", "mink",
		"viper", "python", "condor", "sparrow", "robin", "wren", "finch", "lark",
	}
	
	rand.Seed(time.Now().UnixNano())
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	
	return fmt.Sprintf("%s-%s", adj, noun)
}
