package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"

	"github.com/gwijnja/harvester"
)

// Gzipper compresses a file and presents it to the next processor in the chain.
type Gzipper struct {
	harvester.NextProcessor
}

// Process reads a file and writes the compressed contents to the next processor
func (z *Gzipper) Process(filename string, r io.Reader) error {

	// Create a gzip writer
	buf := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(buf)
	defer gzipWriter.Close()

	// Copy the ctx.Reader to the gzip writer
	slog.Info("Copying to gzip entry", slog.String("filename", filename))
	written, err := harvester.AuditCopy(gzipWriter, r)
	if err != nil {
		return fmt.Errorf("Gzipper.Process(): error copying %s after %d bytes: %s", filename, written, err)
	}
	slog.Info("Copy complete", slog.String("filename", filename), slog.Int64("bytes", written))

	gzipWriter.Close()
	slog.Info("Gzip writer closed", slog.Int("buffersize", buf.Len()))

	slog.Info("Renaming context filename", slog.String("filename", filename+".gz"))
	r = bytes.NewReader(buf.Bytes())
	filename = filename + ".gz"

	slog.Debug("Calling the next processor")
	return z.NextProcessor.Process(filename, r)
}
