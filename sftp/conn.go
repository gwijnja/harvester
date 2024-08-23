package sftp

import (
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Connection struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func (c *Connection) Close() {
	if c.sftpClient != nil {
		c.sftpClient.Close()
	}
	if c.sshClient != nil {
		c.sshClient.Close()
	}
}
