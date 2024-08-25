package local

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gwijnja/harvester"
)

// FileWriter is a connector that receives files for the local filesystem.
type FileWriter struct {
	harvester.NextProcessor
	Transmit string
	ToLoad   string
}

// Process receives a file and writes it to the local filesystem.
func (w *FileWriter) Process(filename string, r io.Reader) error {

	// Create the file in the Transmit directory
	transmitPath := filepath.Join(w.Transmit, filename)
	f, err := os.Create(transmitPath)
	if err != nil {
		return fmt.Errorf("local: Failed to open file %s: %s", transmitPath, err)
	}
	slog.Info("local: Created file", slog.String("path", transmitPath))

	defer func() {
		f.Close()
		slog.Info("local: Closed file", slog.String("path", transmitPath))
	}()

	// Copy the reader to the file
	written, err := harvester.AuditCopy(f, r)
	if err != nil {
		// If the copy fails, close the file and delete it if something was created
		slog.Warn("local: Copy failed, closing and removing the transmit file.")

		f.Close()
		slog.Info("local: Closed file", slog.String("path", transmitPath))

		os.Remove(transmitPath)
		slog.Info("local: Removed file", slog.String("path", transmitPath))

		return err
	}
	slog.Info("local: Copied the file", slog.String("path", transmitPath), slog.Int64("written", written))

	// Move the file from Transmit to ToLoad
	toLoadPath := fmt.Sprintf("%s/%s", w.ToLoad, filename)
	err = os.Rename(transmitPath, toLoadPath)
	if err != nil {
		slog.Warn("local: Move to ToLoad failed, closing and removing the transmit file.")

		f.Close()
		slog.Info("local: Closed file", slog.String("path", transmitPath))

		os.Remove(transmitPath)
		slog.Info("local: Removed file", slog.String("path", transmitPath))

		return fmt.Errorf("local: Failed to move file %s to %s: %s", transmitPath, toLoadPath, err)
	}
	slog.Info("total: Moved file", slog.String("from", transmitPath), slog.String("to", toLoadPath))

	return nil
}
