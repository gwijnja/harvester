package sftp

import (
	"log/slog"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type connection struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func (c *connection) Close() {
	if c.sftpClient != nil {
		c.sftpClient.Close()
		slog.Info("sftp: Closed SFTP client")
	}
	if c.sshClient != nil {
		c.sshClient.Close()
		slog.Info("sftp: Closed SSH connection")
	}
}
