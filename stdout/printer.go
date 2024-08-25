package stdout

import (
	"io"
	"log/slog"
	"strings"

	"github.com/gwijnja/harvester"
)

// Printer prints the contents of a file to stdout
type Printer struct {
	harvester.NextProcessor
}

// Process reads a file and writes the contents to stdout
func (p *Printer) Process(filename string, r io.Reader) error {

	buf := new(strings.Builder)

	_, err := harvester.AuditCopy(buf, r)
	if err != nil {
		return err
	}
	slog.Info("stdout: Copied contents", slog.String("filename", filename), slog.String("contents", buf.String()))

	return nil
}
