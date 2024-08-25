package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"github.com/gwijnja/harvester"
)

// Decompressor extracts a file from a zip archive and presents it to the next processor in the chain.
type Decompressor struct {
	harvester.NextProcessor
}

// Process reads a zip file and writes the contents of the first file to the next processor
func (u *Decompressor) Process(_ string, r io.Reader) error {

	// Copy the contents of the reader to a buffer
	var buf bytes.Buffer
	written, err := harvester.AuditCopy(&buf, r)
	if err != nil {
		return fmt.Errorf("zip: Failed to copy reader to buffer: %s", err)
	}
	slog.Debug("zip: Copied contents to a buffer", slog.Int64("bytes", written))

	// Initialize a zip reader
	reader := bytes.NewReader(buf.Bytes())
	zipReader, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		return fmt.Errorf("zip: Failed to initialize zip reader: %s", err)
	}
	slog.Info("zip: Initialized zip reader")

	// Check if the zip reader contains exactly one file
	if len(zipReader.File) != 1 {
		return fmt.Errorf("zip: Expected only one file in the zip reader, found %d", len(zipReader.File))
	}
	slog.Info("zip: Found one entry in the zip reader", slog.String("filename", zipReader.File[0].Name))

	// Check if the entry is a file, not a directory
	file := zipReader.File[0]
	if file.FileInfo().IsDir() {
		return fmt.Errorf("zip: Expected a file, got a directory: %s", file.Name)
	}

	// Open the file in the zip reader
	readCloser, err := file.Open()
	if err != nil {
		return fmt.Errorf("zip: Failed to open the file in the zip reader: %s", err)
	}
	slog.Debug("zip: Opened the file in the zip reader")

	defer func() {
		readCloser.Close()
		slog.Info("zip: Closed the file in the zip reader", slog.String("filename", file.Name))
	}()

	fileBuf := new(bytes.Buffer)
	if _, err := harvester.AuditCopy(fileBuf, readCloser); err != nil {
		return fmt.Errorf("zip: Failed copying the file into a buffer: %s", err)
	}
	slog.Debug(
		"zip: Extracted the file into a buffer",
		slog.String("filename", file.Name),
		slog.Int64("bytes", int64(fileBuf.Len())),
	)

	r = bytes.NewReader(fileBuf.Bytes())
	return u.NextProcessor.Process(file.Name, r)
}
