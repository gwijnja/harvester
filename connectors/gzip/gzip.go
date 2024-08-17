package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"

	"github.com/gwijnja/harvester"
)

type Gzipper struct {
	harvester.BaseProcessor
}

func (z *Gzipper) Process(ctx *harvester.FileContext) error {

	log.Println("[gzip] Gzipper.Process(): Called for", ctx.Filename)

	// Create a gzip writer
	log.Println("[gzip] Gzipper.Process(): Creating a gzip writer")
	buf := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(buf)
	defer gzipWriter.Close()

	// Copy the ctx.Reader to the gzip writer
	log.Println("[gzip] Gzipper.Process(): Copying the contents from the ctx Reader to the gzip entry")
	written, err := harvester.AuditCopy(gzipWriter, ctx.Reader)
	if err != nil {
		return fmt.Errorf("[gzip] Gzipper.Process(): error copying %s after %d bytes: %s", ctx.Filename, written, err)
	}
	log.Println("[gzip] Gzipper.Process(): Copy complete, copied", written, "bytes")

	log.Println("[gzip] Gzipper.Process(): Closing the gzip archive")
	gzipWriter.Close()

	log.Println("[gzip] Gzipper.Process(): Buffer size after closing:", buf.Len(), "bytes")

	log.Println("[gzip] Gzipper.Process(): Renaming the context filename to", ctx.Filename+".gz")
	ctx.Reader = buf
	ctx.Filename = ctx.Filename + ".gz"

	log.Println("[gzip] Gzipper.Process(): Calling the next processor")
	return z.BaseProcessor.Process(ctx)
}
