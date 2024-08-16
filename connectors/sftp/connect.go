package sftp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func (c *Connector) connect() error {
	log.Printf("[sftp] Connecting to %s@%s:%d\n", c.Username, c.Host, c.Port)

	var auths []ssh.AuthMethod

	// Add password authentication
	auths = addPasswordAuth(auths, c.Password)

	// Add private key authentication
	auths, err := addPrivateKeyAuth(auths, c.PrivateKeyFile, c.Passphrase)
	if err != nil {
		return fmt.Errorf("[sftp] failed to add private key auth: %s", err)
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
				return fmt.Errorf("[sftp] failed to open known_hosts file: %s", err)
			}
			log.Println("[sftp] Known hosts file not found, continuing without host key verification")
		} else {
			config.HostKeyCallback = hostKeyCallback
		}
	}

	// Connect to the SSH server
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	conn, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return fmt.Errorf("[sftp] failed to dial: %s", err)
	}

	log.Println("[sftp] Connected!")

	// Create a new SFTP client
	log.Println("[sftp] Creating SFTP client...")
	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("[sftp] failed to create SFTP client: %s", err)
	}
	log.Println("[sftp] SFTP client created")

	c.client = client
	return nil
}

func addPasswordAuth(auths []ssh.AuthMethod, password string) []ssh.AuthMethod {
	if password != "" {
		return append(auths, ssh.Password(password))
	}
	return auths
}

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

func (c *Connector) Close() error {
	if c.client != nil {
		log.Println("[sftp] Closing connection")
		err := c.client.Close()
		c.client = nil
		return err
	}
	return nil
}

func (c *Connector) Finalize() {
	log.Println("[sftp] Finalizing...")
	c.Close()
}

func (c *Connector) isAlive() bool {
	if c.client == nil {
		return false
	}

	_, err := c.client.ReadDir(c.ToLoad)
	return err == nil
}
