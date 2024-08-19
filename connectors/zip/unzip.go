package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log/slog"

	"github.com/gwijnja/harvester"
)

// Unzipper extracts a file from a zip archive and presents it to the next processor in the chain.
type Unzipper struct {
	harvester.BaseProcessor
}

// Process reads a zip file and writes the contents of the first file to the next processor
func (u *Unzipper) Process(ctx *harvester.FileContext) error {

	slog.Info("Copying the source into a buffer to determine the length")
	var buf bytes.Buffer
	if _, err := harvester.AuditCopy(&buf, ctx.Reader); err != nil {
		return fmt.Errorf("Unzipper.Process(): unable to copy the reader into a buffer: %s", err)
	}
	slog.Debug("Buffer created", slog.Int("buffersize", buf.Len()))

	slog.Debug("Wrapping the buffer with a bytes reader")
	reader := bytes.NewReader(buf.Bytes())

	slog.Info("Creating a zip reader")
	zipReader, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		return fmt.Errorf("Unzipper.Process(): unable to create reader for zip file: %s", err)
	}
	slog.Info("Zip reader created")

	slog.Info("Files found in the zip", slog.Int("count", len(zipReader.File)))
	if len(zipReader.File) != 1 {
		return fmt.Errorf("Unzipper.Process(): expected exactly one file in the zip, but got %d", len(zipReader.File))
	}

	slog.Info("Checking if the first file in the zip is not a directory")
	file := zipReader.File[0]
	if file.FileInfo().IsDir() {
		return fmt.Errorf("Unzipper.Process(): expected a file in the zip, but got a directory")
	}

	slog.Info("Opening the file in the zip")
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("Unzipper.Process(): unable to open the file in the zip: %s", err)
	}
	defer rc.Close()
	slog.Info("File opened")

	slog.Info("Calling AuditCopy to copy the file into a buffer")
	fileBuf := new(bytes.Buffer)
	if _, err := harvester.AuditCopy(fileBuf, rc); err != nil {
		return fmt.Errorf("Unzipper.Process(): unable to copy the file into the buffer: %s", err)
	}
	slog.Info("File copied into the buffer", slog.Int("buffersize", fileBuf.Len()))

	slog.Info("Replacing the context filename", slog.String("filename", file.Name))
	ctx.Reader = fileBuf
	ctx.Filename = file.Name

	slog.Info("Calling the next processor")
	return u.BaseProcessor.Process(ctx)
}
