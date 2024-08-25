package local

import (
	"fmt"
	"gwijnja/mft"
	"log"
	"os"
)

// Writer is a connector that receives files for the local filesystem.
type Writer struct {
	mft.BaseProcessor
	Transmit string
	ToLoad   string
}

// Process receives a file and writes it to the local filesystem.
func (r *Writer) Process(ctx *mft.FileContext) error {
	log.Println("[local] Writer.Process(): Called for", ctx.Filename)

	path := r.Transmit + "/" + ctx.Filename
	log.Println("[local] Writer.Process(): Creating", path)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("[local] Writer.Process(): unable to open file %s: %s", path, err)
	}
	defer f.Close()

	// Copy the ctx.Reader to the file
	log.Println("[local] Writer.Process(): Calling AuditCopy")
	written, err := mft.AuditCopy(f, ctx.Reader)
	if err != nil {
		// If the copy fails, close the file and delete it if something was created
		log.Println("[local] Writer.Process(): error copying data after", written, "bytes:", err)
		log.Println("[local] Writer.Process(): closing and removing", path)
		f.Close()
		os.Remove(path)
		return fmt.Errorf("[local] Writer.Process(): error copying %s after %d bytes: %s", path, written, err)
	}
	log.Println("[local] Writer.Process(): Copy complete:", written, "bytes")

	// Move the file from Transmit to ToLoad
	log.Println("[local] Writer.Process(): Moving", path, "to", r.ToLoad+"/"+ctx.Filename)
	err = os.Rename(path, r.ToLoad+"/"+ctx.Filename)
	if err != nil {
		// TODO: if the ToLoad directory does not exist, the error is wait too long and confusing.
		log.Println("[local] Writer.Process(): error moving", path, ":", err)
		log.Println("[local] Writer.Process(): closing and removing", path)
		f.Close()
		os.Remove(path)
		// TODO: dubbele errors, moet eigenlijk niet he...
		return fmt.Errorf("[local] Writer.Process(): error moving %s: %s", path, err)
	}

	return nil
}
