package ftp

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/gwijnja/harvester"
	"github.com/jlaffaye/ftp"
)

type FileWriter struct {
	FtpConnector
	Transmit string
	ToLoad   string
}

func (w *FileWriter) SetNext(next harvester.FileWriter) {}

func (w *FileWriter) Process(filename string, r io.Reader) error {
	conn, err := w.connect()
	if err != nil {
		return err
	}
	defer func() {
		slog.Debug("ftp: closing connection")
		conn.Quit()
	}()

	// Set the transfer type to binary
	slog.Debug("Writer: setting transfer type to binary")
	err = conn.Type(ftp.TransferTypeBinary)
	if err != nil {
		return fmt.Errorf("writer: error setting transfer type to binary: %s", err)
	}

	transmitPath := filepath.Join(w.Transmit, filename)
	err = conn.Stor(transmitPath, r)
	if err != nil {
		return fmt.Errorf("writer: error storing file %s: %s", transmitPath, err)
	}

	// Move the file from Transmit to ToLoad
	toLoadPath := filepath.Join(w.ToLoad, filename)
	slog.Info("Writer: renaming file", slog.String("from", transmitPath), slog.String("to", toLoadPath))
	err = conn.Rename(transmitPath, toLoadPath)
	if err != nil {
		return fmt.Errorf("writer: error renaming file %s to %s: %s", transmitPath, toLoadPath, err)
	}

	return nil
}
