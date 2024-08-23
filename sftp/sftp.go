package sftp

import (
	"fmt"
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
}

/*
func (c *Connector) reconnectIfNeeded() error {
	if c.client == nil {
		slog.Info("SFTP client is not connected, connecting")
		return c.connect()
	}

	if !c.isAlive() {
		slog.Info("SFTP client is dead, reconnecting")
		return c.connect()
	}

	return nil
}

func (c *Connector) isAlive() bool {
	if c.client == nil {
		return false
	}
	_, err := c.client.Stat("/")
	return err == nil
}
*/

// connect establishes a connection to the SFTP server
func (c *Connector) connect() (*Connection, error) {

	var auths []ssh.AuthMethod

	// Add password authentication
	auths = addPasswordAuth(auths, c.Password)

	// Add private key authentication
	auths, err := addPrivateKeyAuth(auths, c.PrivateKeyFile, c.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("sftp: failed to add private key auth: %s", err)
	}

	// Create a new SSH client
	config := ssh.ClientConfig{
		User:            c.Username,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// HostKeyAlgorithms: []string{ssh.KeyAlgoDSA},
	}

	// Overwrite the HostKeyCallback if FailIfHostKeyChanged is set
	if c.FailIfHostKeyChanged {
		path := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
		hostKeyCallback, err := knownhosts.New(path)
		if err != nil {
			if c.FailIfHostKeyNotFound {
				return nil, fmt.Errorf("sftp: failed to open known_hosts file: %s", err)
			}
			slog.Warn("sftp: Known hosts file not found, continuing without host key verification")
		} else {
			config.HostKeyCallback = hostKeyCallback
		}
	}

	// Connect to the SSH server
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	slog.Info("sftp: Dialing", slog.String("address", addr))
	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return nil, fmt.Errorf("sftp: failed to dial: %s", err)
	}

	// Create a new SFTP client
	slog.Info("sftp: Connected, creating SFTP client")
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("sftp: failed to create SFTP client: %s", err)
	}

	// Return both, because both must be closed at the same time
	return &Connection{
		sshClient:  sshClient,
		sftpClient: sftpClient,
	}, nil
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
