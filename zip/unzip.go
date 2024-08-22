package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"github.com/gwijnja/harvester"
)

// Unzipper extracts a file from a zip archive and presents it to the next processor in the chain.
type Unzipper struct {
	harvester.NextProcessor
}

// Process reads a zip file and writes the contents of the first file to the next processor
func (u *Unzipper) Process(filename string, r io.Reader) error {

	slog.Info("Copying the source into a buffer to determine the length")
	var buf bytes.Buffer
	if _, err := harvester.AuditCopy(&buf, r); err != nil {
		return fmt.Errorf("Unzipper.Process(): unable to copy the reader into a buffer: %s", err)
	}

	slog.Debug("Buffer created, wrapping it in a bytes reader.", slog.Int("buffersize", buf.Len()))
	reader := bytes.NewReader(buf.Bytes())

	slog.Info("Creating a zip reader")
	zipReader, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		return fmt.Errorf("Unzipper.Process(): unable to create reader for zip file: %s", err)
	}

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
	readCloser, err := file.Open()
	if err != nil {
		return fmt.Errorf("Unzipper.Process(): unable to open the file in the zip: %s", err)
	}
	defer readCloser.Close()

	slog.Info("Calling AuditCopy to copy the file into a buffer")
	fileBuf := new(bytes.Buffer)
	if _, err := harvester.AuditCopy(fileBuf, readCloser); err != nil {
		return fmt.Errorf("Unzipper.Process(): unable to copy the file into the buffer: %s", err)
	}

	slog.Info("Replacing the context filename", slog.String("filename", file.Name))
	r = bytes.NewReader(fileBuf.Bytes())
	filename = file.Name

	slog.Info("Calling the next processor")
	return u.NextProcessor.Process(filename, r)
}
