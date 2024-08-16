package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"gwijnja/mft"
	"log"
)

type Unzipper struct {
	mft.BaseProcessor
}

func (u *Unzipper) Process(ctx *mft.FileContext) error {
	log.Println("[zip] Unzipper.Process(): Called for", ctx.Filename)

	log.Println("[zip] Unzipper.Process(): Copying the io.Reader into a buffer, because the zip reader needs to known the data length. Calling AuditCopy.")
	var buf bytes.Buffer
	if _, err := mft.AuditCopy(&buf, ctx.Reader); err != nil {
		return fmt.Errorf("[zip] Unzipper.Process(): unable to copy the reader into a buffer: %s", err)
	}

	log.Println("[zip] Unzipper.Process(): Wrapping the buffer with a bytes reader")
	reader := bytes.NewReader(buf.Bytes())

	log.Println("[zip] Unzipper.Process(): Creating a zip reader")
	zipReader, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		return fmt.Errorf("[zip] Unzipper.Process(): unable to create reader for zip file: %s", err)
	}

	log.Println("[zip] Unzipper.Process():", len(zipReader.File), "files found in the zip file")
	if len(zipReader.File) != 1 {
		return fmt.Errorf("[zip] Unzipper.Process(): expected exactly one file in the zip, but got %d", len(zipReader.File))
	}

	log.Println("[zip] Unzipper.Process(): Opening the first file in the zip, checking if it's not a directory")
	file := zipReader.File[0]
	if file.FileInfo().IsDir() {
		return fmt.Errorf("[zip] Unzipper.Process(): expected a file in the zip, but got a directory")
	}

	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("[zip] Unzipper.Process(): unable to open the file in the zip: %s", err)
	}
	defer rc.Close()

	log.Println("[zip] Unzipper.Process(): Reading the file into memory, calling AuditCopy")
	fileBuf := new(bytes.Buffer)
	if _, err := mft.AuditCopy(fileBuf, rc); err != nil {
		return fmt.Errorf("[zip] Unzipper.Process(): unable to copy the file into the buffer: %s", err)
	}

	log.Println("[zip] Unzipper.Process(): Replacing the context filename with", file.Name)
	ctx.Reader = fileBuf
	ctx.Filename = file.Name

	log.Println("[zip] Unzipper.Process(): Calling the next processor")
	return u.BaseProcessor.Process(ctx)
}
