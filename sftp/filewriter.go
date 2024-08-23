package sftp

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/gwijnja/harvester"
)

type FileWriter struct {
	Connector
	Transmit string
	ToLoad   string
}

func (w *FileWriter) SetNext(next harvester.FileWriter) {}

func (w *FileWriter) Process(filename string, r io.Reader) error {
	w.Connector.reconnectIfNeeded()

	// Open the file to write to
	transmitPath := filepath.Join(w.Transmit, filename)
	slog.Info("sftp: Opening file", slog.String("filename", transmitPath))
	f, err := w.Connector.client.Create(transmitPath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %s", transmitPath, err)
	}
	defer f.Close()

	// Call AuditCopy to write the file
	slog.Info("sftp: Calling AuditCopy to write the file")
	_, err = harvester.AuditCopy(f, r)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %s", filename, err)
	}

	// Move the file to the toload directory
	toLoadPath := filepath.Join(w.ToLoad, filename)
	slog.Info("sftp: Moving file", slog.String("from", transmitPath), slog.String("to", toLoadPath))
	err = w.Connector.client.Rename(transmitPath, toLoadPath) // TODO: Use PosixRename?
	if err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %s", transmitPath, toLoadPath, err)
	}

	return nil
}
