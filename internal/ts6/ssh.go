package ts6

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"ts6-viewer/internal/config"

	"golang.org/x/crypto/ssh"
)

// sshReadBufferSize is the size of the bufio.Reader buffer used for reading
// ServerQuery responses. The default 4096 is too small for responses like
// channellist or clientlist on servers with many channels/clients.
const sshReadBufferSize = 65536

// ServerEvent is one ServerQuery event notification (a line starting with
// "notify"). Name is the event name (e.g. "notifycliententerview"), Fields
// holds the key=value pairs in the order they appear on the line, Raw is
// the full line. Callbacks may rely on Name; the field names are parsed
// best-effort and should be treated as informational.
type ServerEvent struct {
	Name   string
	Fields map[string]string
	Raw    string
}

// execResult is the outcome of one Exec: either a non-zero "error id=" line
// (err set) or a clean "error id=0" terminator. The data lines are drained
// from the pending's lines channel by the caller.
type execResult struct {
	err error
}

// pending represents one in-flight command. ServerQuery is synchronous and
// responses carry no request IDs, so at most one command may be in flight:
// data lines go into lines, the terminating "error id=" line into result.
type pending struct {
	lines  chan string
	result chan execResult
}

// SSHClient is a persistent SSH ServerQuery connection. A single always-on
// reader goroutine owns the stdout stream: it demultiplexes the byte
// stream into command responses and asynchronous event notifications
// (notify* lines), so events are observed in real time while commands keep
// working over the same session.
//
// Only that one goroutine ever reads from the bufio.Reader. This is not
// optional: the reader is a shared stateful buffer, and a second
// concurrent reader (for example a keepalive racing the sampler) corrupts
// the buffered state and shifts whole lines between responses.
type SSHClient struct {
	cfg      *config.Config
	serverID string

	ssh     *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	reader  *bufio.Reader // read exclusively by the readLoop goroutine

	mu               sync.Mutex // guards pending, onEvent, eventsRegistered, dead, readErr
	pending          *pending   // the in-flight command, if any
	onEvent          func(ServerEvent)
	eventsRegistered bool // servernotifyregister was sent on this connection
	dead             bool // reader loop exited (connection lost or closed)
	readErr          error // reason the reader loop exited, if any

	done  chan struct{}
	once  sync.Once
	execMu sync.Mutex // serializes Exec across callers (viewer, sampler, keepalive)
}

var (
	globalClient *SSHClient
	globalMu     sync.Mutex

	floodWaitRegex = regexp.MustCompile(`wait (\d+)ms`)
	errIDRegex     = regexp.MustCompile(`^error id=(\d+)`)
)

// GetPersistentClient returns a singleton SSH connection, creating it on
// first use. If the previous connection died (server dropped it, network
// blip) a new one is established transparently.
func GetPersistentClient(cfg *config.Config, serverID string) (*SSHClient, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	if c := globalClient; c != nil && !c.IsClosed() && !c.IsDead() {
		log.Println("[SSH] Reusing existing persistent connection")
		return c, nil
	}

	if globalClient != nil {
		log.Println("[SSH] Previous persistent connection is dead, reconnecting")
		globalClient.Close()
		globalClient = nil
	}

	log.Println("[SSH] Creating new persistent SSH connection")

	client, err := newSSHClient(cfg, serverID)
	if err != nil {
		log.Printf("[SSH] Connection creation failed: %v\n", err)
		return nil, err
	}

	globalClient = client
	log.Println("[SSH] Persistent SSH connection established")

	return globalClient, nil
}

// newSSHClient establishes a fresh SSH ServerQuery connection: TCP dial,
// SSH handshake, shell session, welcome banner, then login and virtual
// server selection. The welcome banner is the only part read before the
// reader loop starts (no other reader exists yet, so that is safe); from
// login onwards every response is read by the reader loop.
func newSSHClient(cfg *config.Config, serverID string) (*SSHClient, error) {
	c, err := newSSHClientBase(cfg)
	if err != nil {
		return nil, err
	}

	c.cfg = cfg
	c.serverID = serverID
	go c.readLoop()

	// Login and server selection now flow through the reader loop like any
	// other command, so exactly one goroutine touches the reader.
	loginCmd := fmt.Sprintf("login %s %s", cfg.Teamspeak6.User, cfg.Teamspeak6.Password)
	if raw, err := c.Exec(loginCmd); err != nil {
		c.Close()
		log.Printf("[SSH] Login failed: %v (server: %s)\n", err, raw)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	if raw, err := c.Exec("use " + serverID); err != nil {
		c.Close()
		log.Printf("[SSH] Virtual server selection failed: %v (server: %s)\n", err, raw)
		return nil, fmt.Errorf("virtual server selection failed: %w", err)
	}

	log.Printf("[SSH] Login successful to %s\n", net.JoinHostPort(cfg.Teamspeak6.Host, cfg.Teamspeak6.Port))

	go c.keepAlive()

	return c, nil
}

// newSSHClientBase creates a raw SSH connection and reads the welcome
// banner. No goroutines are started here, so the banner read cannot race.
func newSSHClientBase(cfg *config.Config) (*SSHClient, error) {
	host := cfg.Teamspeak6.Host
	port := cfg.Teamspeak6.Port
	user := cfg.Teamspeak6.User
	password := cfg.Teamspeak6.Password

	addr := net.JoinHostPort(host, port)

	log.Printf("[SSH] Connecting to %s\n", addr)

	sshConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	rawConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		log.Printf("[SSH] TCP dial failed: %v\n", err)
		return nil, err
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(rawConn, addr, sshConfig)
	if err != nil {
		rawConn.Close()
		log.Printf("[SSH] SSH handshake failed: %v\n", err)
		return nil, err
	}

	client := ssh.NewClient(sshConn, chans, reqs)

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, err
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		client.Close()
		return nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		client.Close()
		return nil, err
	}

	if err := session.Shell(); err != nil {
		client.Close()
		return nil, err
	}

	c := &SSHClient{
		ssh:     client,
		session: session,
		stdin:   stdin,
		reader:  bufio.NewReaderSize(stdout, sshReadBufferSize),
		done:    make(chan struct{}),
	}

	log.Println("[SSH] Waiting for welcome message")

	// The banner is "TS3" followed by a "Welcome to the TeamSpeak ..."
	// line; read until the Welcome line is seen (up to 8 lines so a
	// different banner layout cannot hang the client). Note: the first
	// line ("TS3") must NOT stop the loop — breaking on it would leave
	// the Welcome line in the buffer to be swallowed by the login
	// response.
	for i := 0; i < 8; i++ {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("failed to read welcome banner: %w", err)
		}
		if strings.Contains(line, "Welcome") || strings.Contains(line, "Teamspeak") {
			break
		}
	}

	log.Println("[SSH] Welcome banner received")

	return c, nil
}

// readLoop is the single always-on reader goroutine. It demultiplexes the
// ServerQuery byte stream:
//   - data lines            -> the in-flight command's pending.lines
//   - "error id=N ..."      -> the in-flight command's pending.result
//   - "notify* ..."         -> the registered event callback (real-time)
//
// A failed/timed-out Exec clears c.pending, so late lines for that command
// are logged and dropped instead of leaking into the next command's
// response (ServerQuery has no request IDs, so the line stream is the only
// correlation we have).
func (c *SSHClient) readLoop() {
	for {
		select {
		case <-c.done:
			log.Println("[SSH] Reader loop stopped (connection closed)")
			return
		default:
		}

		line, err := c.reader.ReadString('\n')
		if err != nil {
			log.Printf("[SSH] Connection lost: %v\n", err)

			c.mu.Lock()
			if p := c.pending; p != nil {
				c.pending = nil
				p.result <- execResult{err: fmt.Errorf("connection lost: %v", err)}
			}
			c.dead = true
			c.readErr = err
			c.mu.Unlock()
			return
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "error id=") {
			c.mu.Lock()
			p := c.pending
			c.mu.Unlock()

			if p == nil {
				log.Printf("[SSH] Stray response line (no command in flight): %s\n", trimmed)
				continue
			}

			if trimmed != "error id=0 msg=ok" {
				p.result <- execResult{err: fmt.Errorf("%s", trimmed)}
			} else {
				p.result <- execResult{}
			}
			continue
		}

		if strings.HasPrefix(trimmed, "notify") {
			c.handleEvent(trimmed)
			continue
		}

		c.mu.Lock()
		p := c.pending
		c.mu.Unlock()

		if p == nil {
			log.Printf("[SSH] Stray data line (no command in flight): %s\n", trimmed)
			continue
		}

		select {
		case p.lines <- trimmed:
		default:
			// Lines buffer full (caller timed out and is draining or gave
			// up); drop the line rather than block the reader loop.
			log.Printf("[SSH] Dropping line for timed-out command: %s\n", trimmed)
		}
	}
}

// handleEvent parses a notify* line into a ServerEvent and invokes the
// registered callback. It is called from the reader loop; the callback must
// not block for long (the http layer coalesces with a non-blocking enqueue).
func (c *SSHClient) handleEvent(raw string) {
	ev := ServerEvent{Raw: raw, Fields: map[string]string{}}

	if sp := strings.IndexByte(raw, ' '); sp > 0 {
		ev.Name = raw[:sp]
		for _, f := range strings.Fields(raw[sp+1:]) {
			if i := strings.IndexByte(f, '='); i > 0 {
				ev.Fields[f[:i]] = f[i+1:]
			}
		}
	} else {
		ev.Name = raw
	}

	c.mu.Lock()
	cb := c.onEvent
	c.mu.Unlock()

	log.Printf("[SSH] Event: %s\n", raw)

	if cb != nil {
		cb(ev)
	}
}

// SetEventCallback registers the callback that receives real-time ServerQuery
// event notifications (notify* lines) and makes sure this connection is
// subscribed for server events ("servernotifyregister event=server"). The
// subscription is sent once per connection and is re-sent automatically
// after every reconnect, so the caller can invoke this once, at startup.
func (c *SSHClient) SetEventCallback(cb func(ServerEvent)) {
	if cb == nil {
		return
	}

	c.mu.Lock()
	c.onEvent = cb
	register := !c.eventsRegistered
	c.mu.Unlock()

	if !register {
		return
	}

	raw, err := c.Exec("servernotifyregister event=server")
	if err != nil {
		log.Printf("[SSH] Event registration failed: %v (server: %s)\n", err, raw)
		return
	}

	c.mu.Lock()
	c.eventsRegistered = true
	c.mu.Unlock()
	log.Println("[SSH] Subscribed to server events (servernotifyregister event=server)")
}

// keepAliveInterval is how often the connection sends a lightweight command
// to prove liveness. The server drops idle ServerQuery sessions after about
// five minutes (measured live), so this must stay well below that.
const keepAliveInterval = 20 * time.Second

// keepAlive sends periodic "version" commands until the connection dies or
// is closed. The version response is consumed by the reader loop like any
// other command, so no extra reading is needed here.
func (c *SSHClient) keepAlive() {
	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			log.Println("[SSH] Keepalive stopped (connection closed)")
			return
		case <-ticker.C:
			if c.IsClosed() || c.IsDead() {
				return
			}
			raw, err := c.Exec("version")
			if err != nil {
				log.Printf("[SSH] Keepalive failed: %v (server: %s). Attempting reconnect\n", err, raw)
				c.Close()
				if _, rerr := reconnectTo(c); rerr != nil {
					log.Printf("[SSH] Reconnect after keepalive failure: %v\n", rerr)
				}
				return
			}
		}
	}
}

// Exec executes a ServerQuery command and waits for its response. The
// contract is unchanged from the pre-event client: the returned string is
// every response line (including the final "error id=0 msg=ok") joined
// with "\n"; a non-zero "error id" line is returned both in the string and
// as an error.
//
// ServerQuery is synchronous and responses carry no request IDs, so only
// one command may be in flight per connection: execMu serializes all
// callers (viewer, sampler, keepalive, event registration).
func (c *SSHClient) Exec(cmd string) (string, error) {
	c.execMu.Lock()
	defer c.execMu.Unlock()

	log.Printf("[SSH] Executing command: %s\n", redactSecret(cmd))
	return c.execSafe(cmd)
}

// redactSecret masks the password in a "login user password" command so
// credentials never reach the log.
func redactSecret(cmd string) string {
	if strings.HasPrefix(cmd, "login ") {
		return "login <user> <password>"
	}
	return cmd
}

// exec sends one command and waits for its response with a timeout. It
// includes panic recovery to handle unexpected bufio errors gracefully
// instead of crashing the process.
func (c *SSHClient) exec(cmd string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[SSH] Recovered from panic during exec(%q): %v\n", cmd, r)
			err = fmt.Errorf("panic during command execution: %v", r)
			result = ""
		}
	}()

	c.mu.Lock()
	if c.dead {
		c.mu.Unlock()
		return "", fmt.Errorf("connection lost: %v", c.readErr)
	}
	p := &pending{
		lines:  make(chan string, 4096),
		result: make(chan execResult, 1),
	}
	c.pending = p
	c.mu.Unlock()

	cleanup := func() {
		c.mu.Lock()
		if c.pending == p {
			c.pending = nil
		}
		c.mu.Unlock()
	}

	if _, err := c.stdin.Write([]byte(cmd + "\n")); err != nil {
		cleanup()
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	var r execResult
	var got bool
	select {
	case r = <-p.result:
		got = true
	case <-time.After(15 * time.Second):
	}

	// Drain whatever the reader delivered (or is still delivering) so those
	// lines cannot leak into the next command's response; then clear the
	// pending slot so the reader stops routing lines to this command.
	lines := drainPending(p)
	cleanup()

	if !got {
		return "", fmt.Errorf("read timeout: no response within 15s")
	}
	if r.err != nil {
		// Non-zero "error id" line: return both the data lines (flood
		// backoff parsing) and the error, exactly like the old client did.
		raw := strings.Join(lines, "\n")
		if raw != "" {
			raw += "\n" + r.err.Error()
		} else {
			raw = r.err.Error()
		}
		return raw, fmt.Errorf("%s", r.err.Error())
	}

	// Preserve the old contract: raw includes every response line, the
	// terminating "error id=0 msg=ok" included.
	raw := strings.Join(lines, "\n")
	if raw != "" {
		raw += "\nerror id=0 msg=ok"
	} else {
		raw = "error id=0 msg=ok"
	}
	return raw, nil
}

// drainPending empties the pending's lines channel. The reader loop may
// still be appending for a while (it captured the pending pointer before
// this call), so drain in non-blocking rounds until quiet.
func drainPending(p *pending) []string {
	var lines []string
	for i := 0; i < 1000; i++ {
		select {
		case l := <-p.lines:
			lines = append(lines, l)
		default:
			// One quiet round means the reader is done delivering.
			return lines
		}
	}
	return lines
}

// execSafe handles flood and reconnect logic around exec.
func (c *SSHClient) execSafe(cmd string) (string, error) {
	const (
		maxFloodRetries = 5
		maxReconnects   = 2
		maxWaitMs       = 10000
		jitterMs        = 250
	)

	floodRetries := 0
	reconnects := 0
	active := c

	for {
		raw, err := active.exec(cmd)
		if err == nil {
			return raw, nil
		}

		if m := errIDRegex.FindStringSubmatch(err.Error()); len(m) == 2 {
			id, _ := strconv.Atoi(m[1])
			if id == 524 {
				wait := 1000
				if match := floodWaitRegex.FindStringSubmatch(err.Error()); len(match) == 2 {
					if ms, convErr := strconv.Atoi(match[1]); convErr == nil {
						wait = ms
					}
				}

				backoff := wait * (1 << floodRetries)
				if backoff > maxWaitMs {
					backoff = maxWaitMs
				}

				jitter := rand.Intn(jitterMs + 1)
				sleepMs := backoff + jitter

				log.Printf("[SSH] Flood detected. Backing off %d ms\n", sleepMs)

				time.Sleep(time.Duration(sleepMs) * time.Millisecond)

				floodRetries++
				if floodRetries >= maxFloodRetries {
					if reconnects >= maxReconnects {
						return "", fmt.Errorf("max flood retries reached: %w", err)
					}
					log.Println("[SSH] Flood retry limit reached. Reconnecting")
					active.Close()
					time.Sleep(300 * time.Millisecond)
					active, err = reconnectTo(active)
					if err != nil {
						return "", err
					}
					floodRetries = 0
					reconnects++
				}
				continue
			}
			return raw, err
		}

		if isConnectionError(err) {
			if reconnects >= maxReconnects {
				return "", err
			}
			log.Printf("[SSH] Connection error detected: %v. Reconnecting\n", err)
			active.Close()
			time.Sleep(300 * time.Millisecond)
			active, err = reconnectTo(active)
			if err != nil {
				return "", err
			}
			reconnects++
			continue
		}

		return "", err
	}
}

// reconnectTo establishes a fresh connection (via the singleton path) that
// inherits the dying client's event subscription, and makes the global
// pointer refer to it. It returns the new client.
func reconnectTo(dead *SSHClient) (*SSHClient, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	log.Println("[SSH] Attempting reconnect")

	if dead.cfg == nil {
		return nil, fmt.Errorf("missing configuration for reconnect")
	}

	c, err := newSSHClient(dead.cfg, dead.serverID)
	if err != nil {
		log.Printf("[SSH] Reconnect failed: %v\n", err)
		return nil, err
	}

	// Inherit the event subscription (callback + registration) so real-time
	// statistics capture survives reconnects.
	dead.mu.Lock()
	cb := dead.onEvent
	registered := dead.eventsRegistered
	dead.mu.Unlock()
	if cb != nil && registered {
		c.mu.Lock()
		c.onEvent = cb
		c.mu.Unlock()
		if raw, err := c.Exec("servernotifyregister event=server"); err != nil {
			log.Printf("[SSH] Re-registering events after reconnect failed: %v (server: %s)\n", err, raw)
		} else {
			c.mu.Lock()
			c.eventsRegistered = true
			c.mu.Unlock()
			log.Println("[SSH] Event subscription restored after reconnect")
		}
	}

	globalClient = c
	log.Println("[SSH] Reconnect successful")

	return c, nil
}

// isConnectionError detects network-level failures.
func isConnectionError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "eof") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "connection lost") ||
		strings.Contains(msg, "panic during command execution") ||
		strings.Contains(msg, "read timeout")
}

// IsClosed checks whether the client is closed.
func (c *SSHClient) IsClosed() bool {
	return c == nil || c.ssh == nil
}

// IsDead reports whether the server dropped the connection (reader loop
// exited with an error). A dead client must be replaced, not reused.
func (c *SSHClient) IsDead() bool {
	if c == nil {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dead
}

// Close terminates the SSH session and signals the reader and keepalive
// goroutines to stop.
func (c *SSHClient) Close() {
	log.Println("[SSH] Closing SSH connection")

	c.once.Do(func() {
		close(c.done)
	})

	if c.session != nil {
		_ = c.session.Close()
		c.session = nil
	}
	if c.ssh != nil {
		_ = c.ssh.Close()
		c.ssh = nil
	}
}
