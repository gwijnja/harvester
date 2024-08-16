package stdout

import (
	"gwijnja/mft"
	"log"
	"strings"
)

type Receiver struct {
	mft.BaseProcessor
}

func (r *Receiver) Process(ctx *mft.FileContext) error {
	log.Println("Stdout conn: Processing", ctx.Filename)

	buf := new(strings.Builder)

	_, err := mft.AuditCopy(buf, ctx.Reader)
	if err != nil {
		return err
	}
	log.Println("Stdout conn: Contents:", buf.String())

	return nil
}
