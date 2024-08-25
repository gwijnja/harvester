package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gwijnja/harvester"
)

// Compressor compresses a file and presents it to the next processor in the chain.
type Compressor struct {
	harvester.NextProcessor
}

// Process reads a file and writes the compressed contents to the next processor
func (z *Compressor) Process(filename string, r io.Reader) error {

	// Create a zip writer
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer func() {
		zipWriter.Close()
		slog.Debug("zip: Zip writer closed")
	}()

	// Create a file in the zip archive
	zipEntryWriter, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("zip: Failed to create file in zip writer for %s: %s", filename, err)
	}
	slog.Info("zip: Zip entry created", slog.String("filename", filename))

	// Copy the ctx.Reader to the zip archive
	written, err := harvester.AuditCopy(zipEntryWriter, r)
	if err != nil {
		return err
	}
	slog.Info("zip: Copied data to zip writer", slog.Int64("bytes", written))

	// Close the zip archive
	zipWriter.Close()
	slog.Debug("zip: Closed zip writer")

	// Rename the filename
	extension := filepath.Ext(filename)
	withoutExtension := strings.TrimSuffix(filename, extension)
	withZipExtension := withoutExtension + ".zip"
	slog.Info("zip: Renamed the context", slog.String("newname", withZipExtension))

	r = bytes.NewReader(buf.Bytes())
	return z.NextProcessor.Process(withZipExtension, r)
}
