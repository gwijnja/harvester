package ftp

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/gwijnja/harvester"
	"github.com/jlaffaye/ftp"
)

type Uploader struct {
	Connector
	Transmit string
	ToLoad   string
}

// SetNext is a no-op for the FileWriter
func (u *Uploader) SetNext(next harvester.FileWriter) {}

// Process writes the file to the FTP server and moves it to the ToLoad directory
func (u *Uploader) Process(filename string, r io.Reader) error {

	// Connect
	conn, err := u.connect()
	if err != nil {
		return err
	}
	defer func() {
		// Close the connection
		conn.Quit()
		slog.Info("ftp: Closed connection")
	}()

	// Set the transfer type to binary
	err = conn.Type(ftp.TransferTypeBinary)
	if err != nil {
		return fmt.Errorf("ftp: Failed to set transfer type to binary: %s", err)
	}
	slog.Debug("ftp: Set transfer type to binary")

	// Store the file in the Transmit directory
	transmitPath := filepath.Join(u.Transmit, filename)
	err = conn.Stor(transmitPath, r)
	if err != nil {
		return fmt.Errorf("ftp: Failed to store file %s: %s", transmitPath, err)
	}
	slog.Info("ftp: Stored file", slog.String("path", transmitPath))

	// Move the file from Transmit to ToLoad
	toLoadPath := filepath.Join(u.ToLoad, filename)
	err = conn.Rename(transmitPath, toLoadPath)
	if err != nil {
		return fmt.Errorf("ftp: Failed to rename file %s to %s: %s", transmitPath, toLoadPath, err)
	}
	slog.Info("ftp: Renamed file", slog.String("from", transmitPath), slog.String("to", toLoadPath))

	return nil
}
