package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"gwijnja/mft"
	"log"
	"strings"
)

type Gunzipper struct {
	mft.BaseProcessor
}

func (z *Gunzipper) Process(ctx *mft.FileContext) error {

	log.Println("[gzip] Gunzipper.Process(): Called for", ctx.Filename)

	// Create a gzip reader
	log.Println("[gzip] Gunzipper.Process(): Creating a gzip reader")
	gzipReader, err := gzip.NewReader(ctx.Reader)
	if err != nil {
		return fmt.Errorf("[gzip] Gunzipper.Process(): error creating gzip reader for %s: %s", ctx.Filename, err)
	}
	defer gzipReader.Close()

	// Copy the gzip reader to a buffer
	buf := new(bytes.Buffer)
	log.Println("[gzip] Gunzipper.Process(): Copying the contents from the gzip reader to the buffer")
	written, err := mft.AuditCopy(buf, gzipReader)
	if err != nil {
		return fmt.Errorf("[gzip] Gunzipper.Process(): error copying %s after %d bytes: %s", ctx.Filename, written, err)
	}
	log.Println("[gzip] Gunzipper.Process(): Copy complete, copied", written, "bytes")

	if strings.HasSuffix(ctx.Filename, ".gz") {
		log.Println("[gzip] Gunzipper.Process(): Removing the .gz suffix from the context filename")
		ctx.Filename = strings.TrimSuffix(ctx.Filename, ".gz")
	}

	ctx.Reader = buf

	log.Println("[gzip] Gunzipper.Process(): Calling the next processor")
	return z.BaseProcessor.Process(ctx)
}
