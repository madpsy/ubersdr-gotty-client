package gottyclient

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/containerd/console"
	"github.com/creack/goselect"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// message types for gotty
const (
	OutputV1         = '0'
	PongV1           = '1'
	SetWindowTitleV1 = '2'
	SetPreferencesV1 = '3'
	SetReconnectV1   = '4'

	InputV1          = '0'
	PingV1           = '1'
	ResizeTerminalV1 = '2'
)

// message types for gotty v2.0
const (
	// Unknown message type, maybe set by a bug
	UnknownOutput = '0'
	// Normal output to the terminal
	Output = '1'
	// Pong to the browser
	Pong = '2'
	// Set window title of the terminal
	SetWindowTitle = '3'
	// Set terminal preference
	SetPreferences = '4'
	// Make terminal to reconnect
	SetReconnect = '5'

	// Unknown message type, maybe sent by a bug
	UnknownInput = '0'
	// User input typically from a keyboard
	Input = '1'
	// Ping to the server
	Ping = '2'
	// Notify that the browser size has been changed
	ResizeTerminal = '3'
)

type gottyMessageType struct {
	output         byte
	pong           byte
	setWindowTitle byte
	setPreferences byte
	setReconnect   byte
	input          byte
	ping           byte
	resizeTerminal byte
}

// GetAuthTokenURL transforms a GoTTY http URL to its AuthToken file URL
func GetAuthTokenURL(httpURL string) (*url.URL, *http.Header, error) {
	header := http.Header{}
	target, err := url.Parse(httpURL)
	if err != nil {
		return nil, nil, err
	}

	target.Path = strings.TrimLeft(target.Path+"auth_token.js", "/")

	user, err := url.PathUnescape(target.User.String())
	if err != nil {
		user = target.User.String()
	}
	if target.User != nil {
		header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user)))
		target.User = nil
	}

	return target, &header, nil
}

// GetURLQuery returns url.query
func GetURLQuery(rawURL string) (url.Values, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return target.Query(), nil
}

// GetWebsocketURL transforms a GoTTY http URL to its WebSocket URL
func GetWebsocketURL(httpURL string) (*url.URL, *http.Header, error) {
	header := http.Header{}
	target, err := url.Parse(httpURL)
	if err != nil {
		return nil, nil, err
	}

	if target.Scheme == "https" {
		target.Scheme = "wss"
	} else {
		target.Scheme = "ws"
	}

	target.Path = strings.TrimLeft(target.Path+"ws", "/")

	user, err := url.PathUnescape(target.User.String())
	if err != nil {
		user = target.User.String()
	}
	if target.User != nil {
		header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user)))
		target.User = nil
	}

	return target, &header, nil
}

type Client struct {
	Dialer          *websocket.Dialer
	Conn            *websocket.Conn
	URL             string
	WriteMutex      *sync.Mutex
	Output          io.Writer
	poison          chan bool
	SkipTLSVerify   bool
	UseProxyFromEnv bool
	Connected       bool
	EscapeKeys      []byte
	V2              bool
	message         *gottyMessageType
	WSOrigin        string
	User            string
	Password        string
	AdminPassword   string
	PathSuffix      string
}

type querySingleType struct {
	AuthToken string `json:"AuthToken"`
	Arguments string `json:"Arguments"`
}

func (c *Client) write(data []byte) error {
	c.WriteMutex.Lock()
	defer c.WriteMutex.Unlock()
	return c.Conn.WriteMessage(websocket.TextMessage, data)
}

// GetAuthToken retrieves an Auth Token from dynamic auth_token.js file
func (c *Client) GetAuthToken() (string, error) {
	target, header, err := GetAuthTokenURL(c.URL)
	if err != nil {
		return "", err
	}
	
	// Add admin password header first (highest priority for proxy authentication)
	if c.AdminPassword != "" {
		header.Add("X-Admin-Password", c.AdminPassword)
	}
	
	// Add basic auth if user is specified
	if c.User != "" {
		basicAuth := c.User + ":" + c.Password
		header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(basicAuth)))
	}

	logrus.Debugf("Fetching auth token auth-token: %q", target.String())
	logrus.Debugf("Request headers: %v", header)
	req, err := http.NewRequest("GET", target.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header = *header
	tr := &http.Transport{}
	if c.SkipTLSVerify {
		conf := &tls.Config{InsecureSkipVerify: true}
		tr.TLSClientConfig = conf
	}
	if c.UseProxyFromEnv {
		tr.Proxy = http.ProxyFromEnvironment
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	switch resp.StatusCode {
	case 200:
		// Everything is OK
	default:
		return "", fmt.Errorf("unknown status code: %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	logrus.Debugf("Auth token response body: %s", string(body))
	
	re := regexp.MustCompile("var gotty_auth_token = '(.*)'")
	output := re.FindStringSubmatch(string(body))
	if len(output) == 0 {
		return "", fmt.Errorf("cannot fetch GoTTY auth-token, please upgrade your GoTTY server")
	}

	authToken := output[1]
	logrus.Debugf("Extracted auth token: %q (length: %d)", authToken, len(authToken))
	return authToken, nil
}

// Connect tries to dial a websocket server
func (c *Client) Connect() error {
	// Retrieve AuthToken
	authToken, err := c.GetAuthToken()
	if err != nil {
		return err
	}
	logrus.Debugf("Auth-token: %q", authToken)

	// Open WebSocket connection
	target, header, err := GetWebsocketURL(c.URL)
	if err != nil {
		return err
	}
	
	// Add admin password header first (highest priority for proxy authentication)
	if c.AdminPassword != "" {
		header.Add("X-Admin-Password", c.AdminPassword)
	}
	
	// Add basic auth if user is specified
	if c.User != "" {
		basicAuth := c.User + ":" + c.Password
		header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(basicAuth)))
	}
	
	if c.WSOrigin != "" {
		header.Add("Origin", c.WSOrigin)
	}
	logrus.Debugf("Connecting to websocket: %q", target.String())
	logrus.Debugf("WebSocket headers: %v", header)
	if c.SkipTLSVerify {
		c.Dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if c.UseProxyFromEnv {
		c.Dialer.Proxy = http.ProxyFromEnvironment
	}
	conn, _, err := c.Dialer.Dial(target.String(), *header)
	if err != nil {
		return err
	}
	c.Conn = conn
	c.Connected = true

	// Initialize message types for gotty BEFORE sending any messages
	c.initMessageType()

	// Pass arguments and auth-token
	query, err := GetURLQuery(c.URL)
	if err != nil {
		return err
	}
	querySingle := querySingleType{
		Arguments: "?" + query.Encode(),
		AuthToken: authToken,
	}
	queryJSON, err := json.Marshal(querySingle)
	if err != nil {
		logrus.Errorf("Failed to parse init message %v", err)
		return err
	}
	// Send Json
	logrus.Debugf("Sending arguments and auth-token: %s", string(queryJSON))
	err = c.write(queryJSON)
	if err != nil {
		return err
	}

	// Small delay to ensure server processes auth message before we start sending other messages
	time.Sleep(100 * time.Millisecond)

	go c.pingLoop()

	return nil
}

// initMessageType initialize message types for gotty
func (c *Client) initMessageType() {
	if c.V2 {
		c.message = &gottyMessageType{
			output:         Output,
			pong:           Pong,
			setWindowTitle: SetWindowTitle,
			setPreferences: SetPreferences,
			setReconnect:   SetReconnect,
			input:          Input,
			ping:           Ping,
			resizeTerminal: ResizeTerminal,
		}
	} else {
		c.message = &gottyMessageType{
			output:         OutputV1,
			pong:           PongV1,
			setWindowTitle: SetWindowTitleV1,
			setPreferences: SetPreferencesV1,
			setReconnect:   SetReconnectV1,
			input:          InputV1,
			ping:           PingV1,
			resizeTerminal: ResizeTerminalV1,
		}
	}
}

func (c *Client) pingLoop() {
	for {
		logrus.Debugf("Sending ping")
		err := c.write([]byte{c.message.ping})
		if err != nil {
			logrus.Warnf("c.write: %v", err)
		}
		time.Sleep(30 * time.Second)
	}
}

// Close will nicely close the dialer
func (c *Client) Close() error {
	return c.Conn.Close()
}

// ExitLoop will kill all goroutines launched by c.Loop()
// ExitLoop() -> wait Loop() -> Close()
func (c *Client) ExitLoop() {
	fname := "ExitLoop"
	openPoison(fname, c.poison)
}

// Loop will look indefinitely for new messages
func (c *Client) Loop() error {

	if !c.Connected {
		err := c.Connect()
		if err != nil {
			return err
		}
	}
	term, err := console.ConsoleFromFile(os.Stdout)
	if err != nil {
		return fmt.Errorf("os.Stdout is not a valid terminal")
	}
	err = term.SetRaw()
	if err != nil {
		return fmt.Errorf("error setting raw terminal: %v", err)
	}
	defer func() {
		_ = term.Reset()
	}()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go c.termsizeLoop(wg)

	wg.Add(1)
	go c.readLoop(wg)

	wg.Add(1)
	go c.writeLoop(wg)

	/* Wait for all of the above goroutines to finish */
	wg.Wait()

	logrus.Debug("Client.Loop() exiting")
	return nil
}

type winsize struct {
	Rows    uint16 `json:"rows"`
	Columns uint16 `json:"columns"`
}

type poisonReason int

const (
	committedSuicide = iota
	killed
)

func openPoison(fname string, poison chan bool) poisonReason {
	logrus.Debug(fname + " suicide")

	/*
	 * The close() may raise panic if multiple goroutines commit suicide at the
	 * same time. Prevent that panic from bubbling up.
	 */
	defer func() {
		if r := recover(); r != nil {
			logrus.Debug("Prevented panic() of simultaneous suicides", r)
		}
	}()

	/* Signal others to die */
	close(poison)
	return committedSuicide
}

func die(fname string, poison chan bool) poisonReason {
	logrus.Debug(fname + " died")

	wasOpen := <-poison
	if wasOpen {
		logrus.Error("ERROR: The channel was open when it wasn't supposed to be")
	}

	return killed
}

func (c *Client) termsizeLoop(wg *sync.WaitGroup) poisonReason {
	defer wg.Done()
	fname := "termsizeLoop"

	ch := make(chan os.Signal, 1)
	notifySignalSIGWINCH(ch)
	defer resetSignalSIGWINCH()

	// Delay first resize to ensure auth message is processed first and terminal is ready
	// Match the web interface delay (100ms) plus extra time for terminal setup
	time.Sleep(200 * time.Millisecond)
	
	// Send initial resize
	if b, err := syscallTIOCGWINSZ(); err != nil {
		// Suppress warning on first attempt - terminal might not be fully ready
		logrus.Debugf("Initial terminal size query failed (expected): %v", err)
	} else {
		if err = c.write(append([]byte{c.message.resizeTerminal}, b...)); err != nil {
			logrus.Warnf("ws.WriteMessage failed: %v", err)
		}
	}
	
	// Handle subsequent resize events
	for {
		select {
		case <-c.poison:
			/* Somebody poisoned the well; die */
			return die(fname, c.poison)
		case <-ch:
			if b, err := syscallTIOCGWINSZ(); err != nil {
				logrus.Warn(err)
			} else {
				if err = c.write(append([]byte{c.message.resizeTerminal}, b...)); err != nil {
					logrus.Warnf("ws.WriteMessage failed: %v", err)
				}
			}
		}
	}
}

type exposeFd interface {
	Fd() uintptr
}

func (c *Client) writeLoop(wg *sync.WaitGroup) poisonReason {
	defer wg.Done()
	fname := "writeLoop"

	buff := make([]byte, 128)

	rdfs := &goselect.FDSet{}
	reader := io.ReadCloser(os.Stdin)

	pr := NewEscapeProxy(reader, c.EscapeKeys)
	defer reader.Close()

	for {
		select {
		case <-c.poison:
			/* Somebody poisoned the well; die */
			return die(fname, c.poison)
		default:
		}

		rdfs.Zero()
		rdfs.Set(reader.(exposeFd).Fd())
		err := goselect.RetrySelect(1, rdfs, nil, nil, 50*time.Millisecond, 3, 50*time.Millisecond)
		if err != nil && err != syscall.EINTR {
			logrus.Debugf(err.Error())
			return openPoison(fname, c.poison)
		}
		if rdfs.IsSet(reader.(exposeFd).Fd()) {
			size, err := pr.Read(buff)

			if err != nil {
				if err == io.EOF {
					// Send EOF to GoTTY

					// Send 'Input' marker, as defined in GoTTY::client_context.go,
					// followed by EOT (a translation of Ctrl-D for terminals)
					err = c.write(append([]byte{c.message.input}, byte(4)))

					if err != nil {
						return openPoison(fname, c.poison)
					}
					continue
				} else {
					return openPoison(fname, c.poison)
				}
			}

			if size <= 0 {
				continue
			}

			data := buff[:size]
			err = c.write(append([]byte{c.message.input}, data...))
			if err != nil {
				return openPoison(fname, c.poison)
			}
		}
	}

}

func (c *Client) readLoop(wg *sync.WaitGroup) poisonReason {
	defer wg.Done()
	fname := "readLoop"

	type MessageNonBlocking struct {
		Data []byte
		Err  error
	}
	msgChan := make(chan MessageNonBlocking)

	for {
		go func() {
			_, data, err := c.Conn.ReadMessage()
			msgChan <- MessageNonBlocking{Data: data, Err: err}
		}()

		select {
		case <-c.poison:
			/* Somebody poisoned the well; die */
			return die(fname, c.poison)
		case msg := <-msgChan:
			if msg.Err != nil {

				if _, ok := msg.Err.(*websocket.CloseError); !ok {
					logrus.Warnf("c.Conn.ReadMessage: %v", msg.Err)
				}
				return openPoison(fname, c.poison)
			}
			if len(msg.Data) == 0 {

				logrus.Warnf("An error has occurred")
				return openPoison(fname, c.poison)
			}
			switch msg.Data[0] {
			case c.message.output: // data
				buf, err := base64.StdEncoding.DecodeString(string(msg.Data[1:]))
				if err != nil {
					logrus.Warnf("Invalid base64 content: %q", msg.Data[1:])
					break
				}
				_, _ = c.Output.Write(buf)
			case c.message.pong: // pong
			case c.message.setWindowTitle: // new title
				newTitle := string(msg.Data[1:])
				_, _ = fmt.Fprintf(c.Output, "\033]0;%s\007", newTitle)
			case c.message.setPreferences: // json prefs
				logrus.Debugf("Received preferences: %s", string(msg.Data[1:]))
			case c.message.setReconnect: // autoreconnect
				var reconnectTimeout int
				if err := json.Unmarshal(msg.Data[1:], &reconnectTimeout); err == nil {
					logrus.Debugf("Server reconnect timeout: %d seconds", reconnectTimeout)
				} else {
					logrus.Debugf("Received reconnect message: %s", string(msg.Data[1:]))
				}
			default:
				logrus.Warnf("Unhandled protocol message: %s", string(msg.Data))
			}
		}
	}
}

// SetOutput changes the output stream
func (c *Client) SetOutput(w io.Writer) {
	c.Output = w
}

// ParseURL parses an URL which may be incomplete and tries to standardize it
func ParseURL(input string) (string, error) {
	parsed, err := url.Parse(input)
	if err != nil {
		return "", err
	}
	switch parsed.Scheme {
	case "http", "https":
		// everything is ok
	default:
		return ParseURL(fmt.Sprintf("http://%s", input))
	}
	return parsed.String(), nil
}

// NewClient returns a GoTTY client object
func NewClient(inputURL string) (*Client, error) {
	parsedURL, err := ParseURL(inputURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		Dialer:     &websocket.Dialer{},
		URL:        parsedURL,
		WriteMutex: &sync.Mutex{},
		Output:     os.Stdout,
		poison:     make(chan bool),
	}, nil
}

// SessionInfo represents information about a tmux session
type SessionInfo struct {
	Name       string `json:"name"`
	WindowName string `json:"window_name,omitempty"`
	Created    string `json:"created"`
	Windows    int    `json:"windows"`
	Attached   bool   `json:"attached"`
	LastActive string `json:"last_active"`
}

// SessionListResponse represents the response for listing sessions
type SessionListResponse struct {
	Sessions []SessionInfo `json:"sessions"`
	Count    int           `json:"count"`
}

// SessionActionResponse represents the response for session actions
type SessionActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Session string `json:"session,omitempty"`
}

// Instance represents a UberSDR instance
type Instance struct {
	ID                    string   `json:"id"`
	Callsign              string   `json:"callsign"`
	Name                  string   `json:"name"`
	Location              string   `json:"location"`
	Latitude              float64  `json:"latitude"`
	Longitude             float64  `json:"longitude"`
	Altitude              int      `json:"altitude"`
	Maidenhead            string   `json:"maidenhead"`
	IsDaylight            bool     `json:"is_daylight"`
	PublicURL             string   `json:"public_url"`
	Version               string   `json:"version"`
	CPUModel              string   `json:"cpu_model"`
	CPUCores              int      `json:"cpu_cores"`
	LoadStatus            string   `json:"load_status"`
	Host                  string   `json:"host"`
	Port                  int      `json:"port"`
	TLS                   bool     `json:"tls,omitempty"`
	MaxClients            int      `json:"max_clients"`
	AvailableClients      int      `json:"available_clients"`
	MaxSessionTime        int      `json:"max_session_time"`
	PublicIQModes         []string `json:"public_iq_modes"`
	SuccessfulCallbacks   int      `json:"successful_callbacks"`
	SNR030MHz             int      `json:"snr_0_30_mhz"`
	SNR1830MHz            int      `json:"snr_1_8_30_mhz"`
	RotatorEnabled        bool     `json:"rotator_enabled,omitempty"`
	RotatorConnected      bool     `json:"rotator_connected,omitempty"`
	RotatorAzimuth        int      `json:"rotator_azimuth"`
	LastReportAgeSeconds  int      `json:"last_report_age_seconds"`
}

// InstanceListResponse represents the response from the instances API
type InstanceListResponse struct {
	Count     int        `json:"count"`
	Instances []Instance `json:"instances"`
}

// ListInstances retrieves the list of available UberSDR instances
func ListInstances() (*InstanceListResponse, error) {
	url := "https://instances.ubersdr.org/api/instances"
	
	logrus.Debugf("Fetching instances list: %q", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list instances: %d %s - %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}

	var instanceList InstanceListResponse
	if err := json.NewDecoder(resp.Body).Decode(&instanceList); err != nil {
		return nil, fmt.Errorf("failed to decode instance list: %v", err)
	}

	return &instanceList, nil
}

// FindInstanceByCallsign finds an instance by its callsign
func FindInstanceByCallsign(callsign string) (*Instance, error) {
	instances, err := ListInstances()
	if err != nil {
		return nil, err
	}

	callsign = strings.ToUpper(callsign)
	for _, instance := range instances.Instances {
		if strings.ToUpper(instance.Callsign) == callsign {
			return &instance, nil
		}
	}

	return nil, fmt.Errorf("no instance found with callsign: %s", callsign)
}

// ListSessions retrieves the list of available tmux sessions
func (c *Client) ListSessions() (*SessionListResponse, error) {
	target, err := url.Parse(c.URL)
	if err != nil {
		return nil, err
	}

	// Build the sessions API URL
	target.Path = strings.TrimRight(target.Path, "/") + "/api/sessions"

	logrus.Debugf("Fetching sessions list: %q", target.String())
	req, err := http.NewRequest("GET", target.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	// Add admin password header first (highest priority for proxy authentication)
	if c.AdminPassword != "" {
		req.Header.Add("X-Admin-Password", c.AdminPassword)
	}
	
	// Add basic auth if user is specified
	if c.User != "" {
		basicAuth := c.User + ":" + c.Password
		req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(basicAuth)))
	}

	// Setup HTTP client
	tr := &http.Transport{}
	if c.SkipTLSVerify {
		conf := &tls.Config{InsecureSkipVerify: true}
		tr.TLSClientConfig = conf
	}
	if c.UseProxyFromEnv {
		tr.Proxy = http.ProxyFromEnvironment
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list sessions: %d %s - %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}

	var sessionList SessionListResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionList); err != nil {
		return nil, fmt.Errorf("failed to decode session list: %v", err)
	}

	return &sessionList, nil
}

// DestroySession destroys a tmux session by name
func (c *Client) DestroySession(sessionName string) (*SessionActionResponse, error) {
	target, err := url.Parse(c.URL)
	if err != nil {
		return nil, err
	}

	// Build the destroy API URL
	target.Path = strings.TrimRight(target.Path, "/") + "/api/sessions/destroy"
	query := target.Query()
	query.Set("name", sessionName)
	target.RawQuery = query.Encode()

	logrus.Debugf("Destroying session: %q", target.String())
	req, err := http.NewRequest("DELETE", target.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	// Add admin password header first (highest priority for proxy authentication)
	if c.AdminPassword != "" {
		req.Header.Add("X-Admin-Password", c.AdminPassword)
	}
	
	// Add basic auth if user is specified
	if c.User != "" {
		basicAuth := c.User + ":" + c.Password
		req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(basicAuth)))
	}

	// Setup HTTP client
	tr := &http.Transport{}
	if c.SkipTLSVerify {
		conf := &tls.Config{InsecureSkipVerify: true}
		tr.TLSClientConfig = conf
	}
	if c.UseProxyFromEnv {
		tr.Proxy = http.ProxyFromEnvironment
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var actionResp SessionActionResponse
	if err := json.NewDecoder(resp.Body).Decode(&actionResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode != 200 {
		return &actionResp, fmt.Errorf("failed to destroy session: %s", actionResp.Message)
	}

	return &actionResp, nil
}
