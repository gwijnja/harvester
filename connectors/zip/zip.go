package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log"

	"github.com/gwijnja/harvester"
)

// Zipper compresses a file and presents it to the next processor in the chain.
type Zipper struct {
	harvester.BaseProcessor
}

// Process reads a file and writes the compressed contents to the next processor
func (z *Zipper) Process(ctx *harvester.FileContext) error {

	log.Println("[zip] Zipper.Process(): Called for", ctx.Filename)

	// Create a zip writer
	log.Println("[zip] Zipper.Process(): Creating a zip writer")
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	// Create a file in the zip archive
	log.Println("[zip] Zipper.Process(): Creating a zip entry for", ctx.Filename)
	zipEntryWriter, err := zipWriter.Create(ctx.Filename)
	if err != nil {
		return fmt.Errorf("[zip] Zipper.Process(): unable to create zip entry for %s: %s", ctx.Filename, err)
	}

	// Copy the ctx.Reader to the zip archive
	log.Println("[zip] Zipper.Process(): Copying the contents from the ctx Reader to the zip entry")
	written, err := harvester.AuditCopy(zipEntryWriter, ctx.Reader)
	if err != nil {
		return fmt.Errorf("[zip] Zipper.Process(): error copying %s after %d bytes: %s", ctx.Filename, written, err)
	}
	log.Println("[zip] Zipper.Process(): Copy complete, copied", written, "bytes to the zip entry")

	log.Println("[zip] Zipper.Process(): Closing the zip archive")
	zipWriter.Close()

	log.Println("[zip] Zipper.Process(): Buffer size after closing:", buf.Len(), "bytes")

	log.Println("[zip] Zipper.Process(): Renaming the context filename to", ctx.Filename+".zip")
	ctx.Reader = buf
	ctx.Filename = ctx.Filename + ".zip"

	log.Println("[zip] Zipper.Process(): Calling the next processor")
	return z.BaseProcessor.Process(ctx)
}
