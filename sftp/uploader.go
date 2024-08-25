package sftp

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/gwijnja/harvester"
)

type Uploader struct {
	Connector
	Transmit string
	ToLoad   string
}

func (u *Uploader) SetNext(next harvester.FileWriter) {}

func (u *Uploader) Process(filename string, r io.Reader) error {

	// Connect to the SFTP server
	conn, err := u.Connector.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Open the file to write to
	transmitPath := filepath.Join(u.Transmit, filename)
	f, err := conn.sftpClient.Create(transmitPath)
	if err != nil {
		return fmt.Errorf("sftp: Failed to create remote file %s: %s", transmitPath, err)
	}
	slog.Info("sftp: Opened remote file", slog.String("path", transmitPath))

	defer func() {
		f.Close()
		slog.Info("sftp: Closed remote file", slog.String("path", transmitPath))
	}()

	// Call AuditCopy to write the file
	_, err = harvester.AuditCopy(f, r)
	if err != nil {
		return err
	}
	slog.Info("sftp: Copied file", slog.String("path", transmitPath))

	// Move the file to the toload directory
	toLoadPath := filepath.Join(u.ToLoad, filename)
	err = conn.sftpClient.Rename(transmitPath, toLoadPath)
	if err != nil {
		return fmt.Errorf("sftp: Failed to move file from %s to %s: %s", transmitPath, toLoadPath, err)
	}
	slog.Info("sftp: Moved file", slog.String("from", transmitPath), slog.String("to", toLoadPath))

	return nil
}
