package stdout

import (
	"log"
	"strings"

	"github.com/gwijnja/harvester"
)

type Receiver struct {
	harvester.BaseProcessor
}

func (r *Receiver) Process(ctx *harvester.FileContext) error {
	log.Println("Stdout conn: Processing", ctx.Filename)

	buf := new(strings.Builder)

	_, err := harvester.AuditCopy(buf, ctx.Reader)
	if err != nil {
		return err
	}
	log.Println("Stdout conn: Contents:", buf.String())

	return nil
}
