package stdout

import (
	"log/slog"
	"strings"

	"github.com/gwijnja/harvester"
)

// Printer prints the contents of a file to stdout
type Printer struct {
	harvester.NextProcessor
}

// Process reads a file and writes the contents to stdout
func (r *Printer) Process(ctx *harvester.FileContext) error {

	buf := new(strings.Builder)

	slog.Info("Calling AuditCopy")
	_, err := harvester.AuditCopy(buf, ctx.Reader)
	if err != nil {
		slog.Error("Error while copying", slog.String("filename", ctx.Filename), slog.Any("error", err))
		return err
	}
	slog.Info("Copy complete", slog.String("contents", buf.String()))

	return nil
}
