package stdout

import (
	"log/slog"
	"strings"

	"github.com/gwijnja/harvester"
)

// Receiver prints the contents of a file to stdout
type Receiver struct {
	harvester.NextProcessor
}

// Process reads a file and writes the contents to stdout
func (r *Receiver) Process(ctx *harvester.FileContext) error {

	buf := new(strings.Builder)

	slog.Info("Calling AuditCopy")
	_, err := harvester.AuditCopy(buf, ctx.Reader)
	if err != nil {
		return err
	}
	slog.Info("AuditCopy returned", slog.String("buf", buf.String()))

	return nil
}
