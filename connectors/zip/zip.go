package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gwijnja/harvester"
)

// Zipper compresses a file and presents it to the next processor in the chain.
type Zipper struct {
	harvester.BaseProcessor
}

// Process reads a file and writes the compressed contents to the next processor
func (z *Zipper) Process(ctx *harvester.FileContext) error {

	// Create a zip writer
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	// Create a file in the zip archive
	slog.Debug("Creating a zip entry", slog.String("filename", ctx.Filename))
	zipEntryWriter, err := zipWriter.Create(ctx.Filename)
	if err != nil {
		return fmt.Errorf("Zipper.Process(): unable to create zip entry for %s: %s", ctx.Filename, err)
	}
	slog.Debug("Zip entry created", slog.String("filename", ctx.Filename))

	// Copy the ctx.Reader to the zip archive
	slog.Info("Calling AuditCopy")
	written, err := harvester.AuditCopy(zipEntryWriter, ctx.Reader)
	if err != nil {
		return fmt.Errorf("Zipper.Process(): error copying %s after %d bytes: %s", ctx.Filename, written, err)
	}
	slog.Info("Copy complete", slog.String("filename", ctx.Filename), slog.Int64("written", written))

	// Close the zip archive
	slog.Debug("Closing the zip archive")
	zipWriter.Close()

	slog.Debug("Zip archive closed", slog.Int("buffersize", buf.Len()))

	extension := filepath.Ext(ctx.Filename)
	withoutExtension := strings.TrimSuffix(ctx.Filename, extension)
	withZipExtension := withoutExtension + ".zip"

	slog.Debug("Renaming the context filename", slog.String("newname", withZipExtension))
	ctx.Reader = buf
	ctx.Filename = withZipExtension

	slog.Debug("Calling the next processor")
	return z.BaseProcessor.Process(ctx)
}
