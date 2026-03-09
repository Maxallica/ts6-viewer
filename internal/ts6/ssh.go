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

// SSHClient represents a persistent SSH ServerQuery connection.
type SSHClient struct {
	cfg      *config.Config
	serverID string

	ssh     *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	reader  *bufio.Reader

	mu sync.Mutex // protects command execution and reconnect
}

var (
	globalClient *SSHClient
	globalMu     sync.Mutex

	floodWaitRegex = regexp.MustCompile(`wait (\d+)ms`)
	errIDRegex     = regexp.MustCompile(`^error id=(\d+)`)
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetPersistentClient returns a singleton SSH connection.
func GetPersistentClient(cfg *config.Config, serverID string) (*SSHClient, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalClient != nil && !globalClient.IsClosed() {
		log.Println("[SSH] Reusing existing persistent connection")
		return globalClient, nil
	}

	log.Println("[SSH] Creating new persistent SSH connection")

	client, err := newSSHClientWithUse(cfg, serverID)
	if err != nil {
		log.Printf("[SSH] Connection creation failed: %v\n", err)
		return nil, err
	}

	globalClient = client
	log.Println("[SSH] Persistent SSH connection established")

	return globalClient, nil
}

// newSSHClientWithUse establishes a new SSH connection and selects the server.
func newSSHClientWithUse(cfg *config.Config, serverID string) (*SSHClient, error) {
	client, err := newSSHClientBase(cfg)
	if err != nil {
		return nil, err
	}

	if err := client.Use(serverID); err != nil {
		client.Close()
		return nil, err
	}

	client.cfg = cfg
	client.serverID = serverID

	go client.keepAlive()

	return client, nil
}

// Use selects the virtual server by ID.
func (c *SSHClient) Use(serverID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsClosed() {
		return fmt.Errorf("ssh connection is closed")
	}

	log.Printf("[SSH] Selecting virtual server: %s\n", serverID)

	_, err := c.stdin.Write([]byte(fmt.Sprintf("use %s\n", serverID)))
	if err != nil {
		return fmt.Errorf("failed to send use command: %w", err)
	}

	timeout := time.After(5 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout while waiting for use response")
		default:
			line, err := c.reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read use response: %w", err)
			}

			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "error id=") {
				if line != "error id=0 msg=ok" {
					return fmt.Errorf("use command failed: %s", line)
				}
				log.Println("[SSH] Virtual server selected successfully")
				return nil
			}
		}
	}
}

// newSSHClientBase creates a raw SSH connection and performs login.
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
		reader:  bufio.NewReader(stdout),
	}

	log.Println("[SSH] Waiting for welcome message")

	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			c.Close()
			return nil, err
		}
		if strings.Contains(line, "Welcome") || strings.Contains(line, "TS3") {
			break
		}
	}

	log.Println("[SSH] Sending login command")

	if _, err := c.stdin.Write([]byte(fmt.Sprintf("login %s %s\n", user, password))); err != nil {
		c.Close()
		return nil, err
	}

	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			c.Close()
			return nil, err
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "error id=") {
			if line != "error id=0 msg=ok" {
				c.Close()
				return nil, fmt.Errorf("login failed: %s", line)
			}
			break
		}
	}

	log.Printf("[SSH] Login successful to %s\n", addr)

	return c, nil
}

// keepAlive sends periodic version commands to prevent idle timeout.
func (c *SSHClient) keepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if c.IsClosed() {
			return
		}
		log.Println("[SSH] Sending keepalive ping")
		_, err := c.Exec("version")
		if err != nil {
			log.Printf("[SSH] Keepalive failed: %v. Attempting reconnect\n", err)
			_ = c.reconnect()
		}
	}
}

// Exec executes a ServerQuery command safely.
func (c *SSHClient) Exec(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Printf("[SSH] Executing command: %s\n", cmd)
	return c.execSafe(cmd)
}

// exec sends a raw command and reads the response.
func (c *SSHClient) exec(cmd string) (string, error) {
	_, err := c.stdin.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", err
	}

	var lines []string
	var last string

	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		lines = append(lines, line)
		last = line
		if strings.HasPrefix(line, "error id=") {
			break
		}
	}

	raw := strings.Join(lines, "\n")

	if strings.HasPrefix(last, "error id=") && last != "error id=0 msg=ok" {
		return raw, fmt.Errorf("%s", last)
	}

	return raw, nil
}

// execSafe handles flood and reconnect logic.
func (c *SSHClient) execSafe(cmd string) (string, error) {
	const (
		maxFloodRetries = 5
		maxReconnects   = 2
		maxWaitMs       = 10000
		jitterMs        = 250
	)

	floodRetries := 0
	reconnects := 0

	for {
		raw, err := c.exec(cmd)
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
					c.Close()
					time.Sleep(300 * time.Millisecond)
					_ = c.reconnect()
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
			c.Close()
			time.Sleep(300 * time.Millisecond)
			_ = c.reconnect()
			reconnects++
			continue
		}

		return "", err
	}
}

// reconnect recreates the SSH connection.
func (c *SSHClient) reconnect() error {
	globalMu.Lock()
	defer globalMu.Unlock()

	log.Println("[SSH] Attempting reconnect")

	if c.cfg == nil {
		return fmt.Errorf("missing configuration for reconnect")
	}

	newClient, err := newSSHClientWithUse(c.cfg, c.serverID)
	if err != nil {
		log.Printf("[SSH] Reconnect failed: %v\n", err)
		return err
	}

	globalClient = newClient
	log.Println("[SSH] Reconnect successful")

	return nil
}

// isConnectionError detects network-level failures.
func isConnectionError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "eof") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "use of closed network connection")
}

// IsClosed checks whether the client is closed.
func (c *SSHClient) IsClosed() bool {
	return c == nil || c.ssh == nil
}

// Close terminates the SSH session.
func (c *SSHClient) Close() {
	log.Println("[SSH] Closing SSH connection")

	if c.session != nil {
		_ = c.session.Close()
		c.session = nil
	}
	if c.ssh != nil {
		_ = c.ssh.Close()
		c.ssh = nil
	}
}
