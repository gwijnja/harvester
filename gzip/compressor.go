package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"

	"github.com/gwijnja/harvester"
)

// Compressor compresses a file using gzip and presents it to the next processor in the chain.
type Compressor struct {
	harvester.NextProcessor
}

// Process reads a file and writes the compressed contents to the next processor
func (c *Compressor) Process(filename string, r io.Reader) error {

	// Create a gzip writer
	buf := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(buf)

	defer func() {
		gzipWriter.Close()
		slog.Info("gzip: Closed gzip writer", slog.String("filename", filename))
	}()

	// Copy the ctx.Reader to the gzip writer
	written, err := harvester.AuditCopy(gzipWriter, r)
	if err != nil {
		return fmt.Errorf("gzip: Failed to copy input to gzip writer: %s", err)
	}
	slog.Info("gzip: Copied input to gzip writer", slog.String("filename", filename), slog.Int64("bytes", written))

	gzipWriter.Close()
	slog.Info("gzip: Closed gzip writer", slog.Int("buffersize", buf.Len()))

	r = bytes.NewReader(buf.Bytes())
	filename = filename + ".gz"
	slog.Info("gzip: Renamed context filename", slog.String("newname", filename+".gz"))

	slog.Debug("Calling the next processor")
	return c.NextProcessor.Process(filename, r)
}
