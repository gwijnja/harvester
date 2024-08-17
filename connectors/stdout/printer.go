package stdout

import (
	"log"
	"strings"

	"github.com/gwijnja/harvester"
)

// Printer prints the contents of a file to stdout
type Printer struct {
	harvester.BaseProcessor
}

// Process reads a file and writes the contents to stdout
func (r *Printer) Process(ctx *harvester.FileContext) error {
	log.Println("Stdout conn: Processing", ctx.Filename)

	buf := new(strings.Builder)

	_, err := harvester.AuditCopy(buf, ctx.Reader)
	if err != nil {
		return err
	}
	log.Println("Stdout conn: Contents:", buf.String())

	return nil
}
