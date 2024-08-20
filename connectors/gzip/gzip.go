package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log/slog"

	"github.com/gwijnja/harvester"
)

// Gzipper compresses a file and presents it to the next processor in the chain.
type Gzipper struct {
	harvester.NextProcessor
}

// Process reads a file and writes the compressed contents to the next processor
func (z *Gzipper) Process(ctx *harvester.FileContext) error {

	// Create a gzip writer
	buf := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(buf)
	defer gzipWriter.Close()

	// Copy the ctx.Reader to the gzip writer
	slog.Info("Copying to gzip entry", slog.String("filename", ctx.Filename))
	written, err := harvester.AuditCopy(gzipWriter, ctx.Reader)
	if err != nil {
		return fmt.Errorf("Gzipper.Process(): error copying %s after %d bytes: %s", ctx.Filename, written, err)
	}
	slog.Info("Copy complete", slog.String("filename", ctx.Filename), slog.Int64("bytes", written))

	gzipWriter.Close()
	slog.Info("Gzip writer closed", slog.Int("buffersize", buf.Len()))

	slog.Info("Renaming context filename", slog.String("filename", ctx.Filename+".gz"))
	ctx.Reader = bytes.NewReader(buf.Bytes())
	ctx.Filename = ctx.Filename + ".gz"

	slog.Debug("Calling the next processor")
	return z.NextProcessor.Process(ctx)
}
