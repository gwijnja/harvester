package sftp

import (
	"log/slog"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Connection struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func (c *Connection) Close() {
	if c.sftpClient != nil {
		slog.Info("Closing SFTP client")
		c.sftpClient.Close()
		slog.Info("SFTP client closed")
	}
	if c.sshClient != nil {
		slog.Info("Closing SSH client")
		c.sshClient.Close()
		slog.Info("SSH client closed")
	}
}
