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

	path := filepath.Join(w.Transmit, filename)
	slog.Info("Creating file", slog.String("path", path))

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("[local] Writer.Process(): unable to open file %s: %s", path, err)
	}
	defer f.Close()
	slog.Info("File created", slog.String("path", path))

	// Copy the r to the file
	slog.Info("Calling AuditCopy")
	written, err := harvester.AuditCopy(f, r)
	if err != nil {
		// If the copy fails, close the file and delete it if something was created
		slog.Error("Error while copying", slog.String("path", path), slog.Any("error", err), slog.Int64("written", written))
		slog.Info("Closing and removing", slog.String("path", path))
		f.Close()
		os.Remove(path)
		return fmt.Errorf("[local] Writer.Process(): error copying %s after %d bytes: %s", path, written, err)
	}
	slog.Info("Copy complete", slog.String("path", path), slog.Int64("written", written))

	// Move the file from Transmit to ToLoad
	dest := fmt.Sprintf("%s/%s", w.ToLoad, filename)
	slog.Info("Moving", slog.String("from", path), slog.String("to", dest))
	err = os.Rename(path, dest)
	if err != nil {
		// TODO: if the ToLoad directory does not exist, the error is wait too long and confusing.
		slog.Error("Error while moving", slog.String("from", path), slog.String("to", dest), slog.Any("error", err))
		slog.Info("Closing and removing", slog.String("path", path))
		f.Close()
		os.Remove(path)
		// TODO: dubbele errors, moet eigenlijk niet he...
		return fmt.Errorf("error moving %s: %s", path, err)
	}

	return nil
}
