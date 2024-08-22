package stdout

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/gwijnja/harvester"
)

// Printer prints the contents of a file to stdout
type FileWriter struct {
	harvester.NextProcessor
}

func (w *FileWriter) Write(filename string, r io.Reader) error {
	slog.Info("Writing file", slog.String("filename", filename))
	_, err := io.Copy(os.Stdout, r)
	if err != nil {
		return fmt.Errorf("Writer.Write(): error writing file %s: %s", filename, err)
	}
	return nil
}

// Process reads a file and writes the contents to stdout
func (w *FileWriter) Process(filename string, r io.Reader) error {

	buf := new(strings.Builder)

	slog.Info("Calling AuditCopy")
	_, err := harvester.AuditCopy(buf, r)
	if err != nil {
		slog.Error("Error while copying", slog.String("filename", filename), slog.Any("error", err))
		return err
	}
	slog.Info("Copy complete", slog.String("contents", buf.String()))

	return nil
}
