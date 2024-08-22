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

// Gunzipper decompresses a gzip file and presents it to the next processor in the chain.
type Gunzipper struct {
	harvester.NextProcessor
}

// Process reads a gzip file and writes the uncompressed contents to the next processor
func (z *Gunzipper) Process(filename string, r io.Reader) error {

	slog.Debug("Creating a gzip reader", slog.String("filename", filename))
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("Gunzipper.Process(): error creating gzip reader for %s: %s", filename, err)
	}
	defer gzipReader.Close()
	slog.Debug("Gzip reader created", slog.String("filename", filename))

	buf := new(bytes.Buffer)
	slog.Debug("Calling AuditCopy")
	written, err := harvester.AuditCopy(buf, gzipReader)
	if err != nil {
		return fmt.Errorf("Gunzipper.Process(): error copying %s after %d bytes: %s", filename, written, err)
	}
	slog.Info("Copy complete", slog.String("filename", filename), slog.Int64("written", written))

	// Remove the .gz suffix from the filename
	if strings.HasSuffix(filename, ".gz") {
		filename = strings.TrimSuffix(filename, ".gz")
		slog.Info("Removing the .gz suffix", slog.String("newname", filename))
	}

	r = bytes.NewReader(buf.Bytes())

	slog.Debug("Calling the next processor")
	return z.NextProcessor.Process(filename, r)
}
