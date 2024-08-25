package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/gwijnja/harvester"
)

// Decompressor decompresses a gzip file and presents it to the next processor in the chain.
type Decompressor struct {
	harvester.NextProcessor
}

// Process reads a gzip file and writes the uncompressed contents to the next processor
func (z *Decompressor) Process(filename string, r io.Reader) error {

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip: Failed to create gzip reader for %s: %s", filename, err)
	}
	slog.Debug("Created gzip reader", slog.String("filename", filename))

	defer func() {
		gzipReader.Close()
		slog.Info("Closed gzip reader", slog.String("filename", filename))
	}()

	// Copy the uncompressed contents to a buffer
	buf := new(bytes.Buffer)
	_, err = harvester.AuditCopy(buf, gzipReader)
	if err != nil {
		return fmt.Errorf("gzip: Failed to copy contents from gzip to buffer: %s", err)
	}
	slog.Info("gzip: Copied contents from gzip to buffer")

	// Remove the .gz suffix from the filename
	if strings.HasSuffix(filename, ".gz") {
		filename = strings.TrimSuffix(filename, ".gz")
		slog.Info("Removed .gz suffix", slog.String("newname", filename))
	}

	r = bytes.NewReader(buf.Bytes())
	slog.Debug("Calling the next processor")
	return z.NextProcessor.Process(filename, r)
}
