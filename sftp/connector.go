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

// connect establishes a connection to the SFTP server
func (c *Connector) connect() (*connection, error) {

	var auths []ssh.AuthMethod

	// Add password authentication
	auths = addPasswordAuth(auths, c.Password)

	// Add private key authentication
	auths, err := addPrivateKeyAuth(auths, c.PrivateKeyFile, c.Passphrase)
	if err != nil {
		return nil, err
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
				return nil, fmt.Errorf("sftp: Failed to open known_hosts file: %s", err)
			}
			slog.Warn("sftp: Failed to open known_hosts file, continuing without host key verification", slog.Any("err", err))
		} else {
			config.HostKeyCallback = hostKeyCallback
		}
	}

	// Connect to the SSH server
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return nil, fmt.Errorf("sftp: Failed to dial: %s", err)
	}
	slog.Info("sftp: Connected", slog.String("address", addr), slog.String("username", c.Username))

	// Create a new SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		slog.Error("sftp: Failed to create SFTP client, closing SSH connection")
		sshClient.Close()
		slog.Info("sftp: Closed SSH connection")
		return nil, fmt.Errorf("sftp: Failed to create SFTP client: %s", err)
	}
	slog.Info("sftp: Created SFTP client")

	// Return both, because both must be closed at the same time
	return &connection{
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

	// Return the list of authentication methods if no private key file is provided
	if privateKeyFile == "" {
		return auths, nil
	}

	// Read the private key from the file
	key, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return auths, fmt.Errorf("sftp: Failed to read private key from %s: %s", privateKeyFile, err)
	}

	// Parse the private key
	var signer ssh.Signer
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(key)
	}

	// Return an error if the private key could not be parsed
	if err != nil {
		return auths, fmt.Errorf("sftp: Failed to parse private key from %s: %s", privateKeyFile, err)
	}

	return append(auths, ssh.PublicKeys(signer)), nil
}
