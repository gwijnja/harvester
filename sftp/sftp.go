package sftp

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Connector is a structure that holds the configuration for an SFTP connection.
type Connector struct {
	Host                  string
	Port                  int
	FailIfHostKeyChanged  bool
	FailIfHostKeyNotFound bool
	Username              string
	Password              string
	PrivateKeyFile        string
	Passphrase            string
	client                *sftp.Client
}

func (c *Connector) isAlive() bool {
	if c.client == nil {
		return false
	}
	_, err := c.client.Stat("/")
	return err == nil
}

// connect establishes a connection to the SFTP server
func (c *Connector) connect() error {
	slog.Info("sftp: Connecting", slog.String("host", c.Host), slog.Int("port", c.Port), slog.String("username", c.Username))

	var auths []ssh.AuthMethod

	// Add password authentication
	auths = addPasswordAuth(auths, c.Password)

	// Add private key authentication
	auths, err := addPrivateKeyAuth(auths, c.PrivateKeyFile, c.Passphrase)
	if err != nil {
		return fmt.Errorf("sftp: failed to add private key auth: %s", err)
	}

	// Create a new SSH client
	config := ssh.ClientConfig{
		User:            c.Username,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Overwrite the HostKeyCallback if FailIfHostKeyChanged is set
	if c.FailIfHostKeyChanged {
		path := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
		hostKeyCallback, err := knownhosts.New(path)
		if err != nil {
			if c.FailIfHostKeyNotFound {
				return fmt.Errorf("sftp: failed to open known_hosts file: %s", err)
			}
			slog.Warn("sftp: Known hosts file not found, continuing without host key verification")
		} else {
			config.HostKeyCallback = hostKeyCallback
		}
	}

	// Connect to the SSH server
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	slog.Info("sftp: Dialing", slog.String("address", addr))
	conn, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return fmt.Errorf("sftp: failed to dial: %s", err)
	}

	// Create a new SFTP client
	slog.Info("sftp: Connected, creating SFTP client")
	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("sftp: failed to create SFTP client: %s", err)
	}

	c.client = client
	return nil
}

// addPasswordAuth adds password authentication to the list of authentication methods
func addPasswordAuth(auths []ssh.AuthMethod, password string) []ssh.AuthMethod {
	if password != "" {
		return append(auths, ssh.Password(password))
	}
	return auths
}

// addPrivateKeyAuth adds private key authentication to the list of authentication methods
func addPrivateKeyAuth(auths []ssh.AuthMethod, privateKeyFile string, passphrase string) ([]ssh.AuthMethod, error) {
	if privateKeyFile == "" {
		return auths, nil
	}

	key, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return auths, fmt.Errorf("failed to read private key: %s", err)
	}

	var signer ssh.Signer

	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(key)
	}

	if err != nil {
		return auths, fmt.Errorf("failed to parse private key: %s", err)
	}

	return append(auths, ssh.PublicKeys(signer)), nil
}

// Close closes the SFTP connection
func (c *Connector) Close() error {
	if c.client != nil {
		log.Println("sftp: Closing connection")
		err := c.client.Close()
		c.client = nil
		return err
	}
	return nil
}

// Finalize closes the SFTP connection
func (c *Connector) Finalize() {
	log.Println("sftp: Finalizing...")
	c.Close()
}
