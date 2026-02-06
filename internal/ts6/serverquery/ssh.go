package serverquery

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
	"ts6-viewer/internal/config"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	ssh     *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	reader  *bufio.Reader
}

// NewSSHClient establishes an SSH connection to the TeamSpeak ServerQuery interface.
func NewSSHClient(cfg *config.Config) (*SSHClient, error) {
	host := cfg.Teamspeak6.Host
	port := cfg.Teamspeak6.Port
	user := cfg.Teamspeak6.User
	password := cfg.Teamspeak6.Password

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	// Establish SSH connection
	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh dial failed: %w", err)
	}

	// Create SSH session
	sess, err := conn.NewSession()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ssh session failed: %w", err)
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		conn.Close()
		return nil, err
	}

	stdout, err := sess.StdoutPipe()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Start interactive shell (required for TS6 ServerQuery)
	if err := sess.Shell(); err != nil {
		conn.Close()
		return nil, err
	}

	client := &SSHClient{
		ssh:     conn,
		session: sess,
		stdin:   stdin,
		reader:  bufio.NewReader(stdout),
	}

	// --- Read welcome message with timeout ---
	welcomeTimeout := time.After(5 * time.Second)
	welcomeReceived := false

	for !welcomeReceived {
		select {
		case <-welcomeTimeout:
			client.Close()
			return nil, fmt.Errorf("timeout while waiting for welcome message")

		default:
			line, err := client.reader.ReadString('\n')
			if err != nil {
				client.Close()
				return nil, fmt.Errorf("failed to read welcome message: %w", err)
			}

			line = strings.TrimSpace(line)

			if strings.Contains(line, "Welcome") || strings.Contains(line, "TS3") {
				welcomeReceived = true
			}
		}
	}

	// --- Send ServerQuery login command ---
	_, err = client.stdin.Write([]byte(fmt.Sprintf("login %s %s\n", user, password)))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// --- Read login response with timeout ---
	loginTimeout := time.After(5 * time.Second)

	for {
		select {
		case <-loginTimeout:
			client.Close()
			return nil, fmt.Errorf("timeout while waiting for login response")

		default:
			line, err := client.reader.ReadString('\n')
			if err != nil {
				client.Close()
				return nil, fmt.Errorf("failed to read login response: %w", err)
			}

			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "error id=") {
				if line != "error id=0 msg=ok" {
					client.Close()
					return nil, fmt.Errorf("login error: %s", line)
				}
				return client, nil
			}
		}
	}
}

// Use selects a virtual server by ID (required before running most commands).
func (c *SSHClient) Use(serverID string) error {
	_, err := c.stdin.Write([]byte(fmt.Sprintf("use %s\n", serverID)))
	if err != nil {
		return fmt.Errorf("failed to send use command: %w", err)
	}

	useTimeout := time.After(5 * time.Second)

	for {
		select {
		case <-useTimeout:
			return fmt.Errorf("timeout while waiting for use response")

		default:
			line, err := c.reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read use response: %w", err)
			}

			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "error id=") {
				if line != "error id=0 msg=ok" {
					return fmt.Errorf("use error: %s", line)
				}
				return nil
			}
		}
	}
}

// exec sends a raw ServerQuery command and returns the raw output.
func (c *SSHClient) exec(cmd string) (string, error) {
	_, err := c.stdin.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", err
	}

	var lines []string

	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimSpace(line)
		lines = append(lines, line)

		// ServerQuery always ends with "error id=..."
		if strings.HasPrefix(line, "error id=") {
			break
		}
	}

	return strings.Join(lines, "\n"), nil
}

// Close cleanly shuts down the SSH session and connection.
func (c *SSHClient) Close() {
	if c.session != nil {
		_ = c.session.Close()
	}
	if c.ssh != nil {
		_ = c.ssh.Close()
	}
}
