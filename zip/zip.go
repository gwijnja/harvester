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

// Zipper compresses a file and presents it to the next processor in the chain.
type Zipper struct {
	harvester.NextProcessor
}

// Process reads a file and writes the compressed contents to the next processor
func (z *Zipper) Process(filename string, r io.Reader) error {

	// Create a zip writer
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	// Create a file in the zip archive
	slog.Debug("Creating a zip entry", slog.String("filename", filename))
	zipEntryWriter, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("Zipper.Process(): unable to create zip entry for %s: %s", filename, err)
	}
	slog.Debug("Zip entry created", slog.String("filename", filename))

	// Copy the ctx.Reader to the zip archive
	slog.Info("Calling AuditCopy")
	written, err := harvester.AuditCopy(zipEntryWriter, r)
	if err != nil {
		return fmt.Errorf("Zipper.Process(): error copying %s after %d bytes: %s", filename, written, err)
	}
	slog.Info("Copy complete", slog.String("filename", filename), slog.Int64("written", written))

	// Close the zip archive
	slog.Debug("Closing the zip archive")
	zipWriter.Close()

	slog.Debug("Zip archive closed", slog.Int("buffersize", buf.Len()))

	extension := filepath.Ext(filename)
	withoutExtension := strings.TrimSuffix(filename, extension)
	withZipExtension := withoutExtension + ".zip"

	slog.Debug("Renaming the context filename", slog.String("newname", withZipExtension))
	r = bytes.NewReader(buf.Bytes())
	filename = withZipExtension

	slog.Debug("Calling the next processor")
	return z.NextProcessor.Process(filename, r)
}
