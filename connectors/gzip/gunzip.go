package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gwijnja/harvester"
)

// Gunzipper decompresses a gzip file and presents it to the next processor in the chain.
type Gunzipper struct {
	harvester.BaseProcessor
}

// Process reads a gzip file and writes the uncompressed contents to the next processor
func (z *Gunzipper) Process(ctx *harvester.FileContext) error {

	// Create a gzip reader
	slog.Info("Creating a gzip reader", slog.String("filename", ctx.Filename))
	gzipReader, err := gzip.NewReader(ctx.Reader)
	if err != nil {
		return fmt.Errorf("Gunzipper.Process(): error creating gzip reader for %s: %s", ctx.Filename, err)
	}
	defer gzipReader.Close()
	slog.Info("Gzip reader created", slog.String("filename", ctx.Filename))

	// Copy the gzip reader to a buffer
	buf := new(bytes.Buffer)
	slog.Info("Calling AuditCopy")
	written, err := harvester.AuditCopy(buf, gzipReader)
	if err != nil {
		return fmt.Errorf("Gunzipper.Process(): error copying %s after %d bytes: %s", ctx.Filename, written, err)
	}
	slog.Info("Copy complete", slog.String("filename", ctx.Filename), slog.Int64("written", written))

	// Remove the .gz suffix from the filename
	if strings.HasSuffix(ctx.Filename, ".gz") {
		ctx.Filename = strings.TrimSuffix(ctx.Filename, ".gz")
		slog.Info("Removing the .gz suffix", slog.String("newname", ctx.Filename))
	}

	ctx.Reader = buf

	slog.Debug("Calling the next processor")
	return z.BaseProcessor.Process(ctx)
}
