package harvester

import (
	"fmt"
	"io"
	"log/slog"
)

type Splitter struct {
	NextProcessor
	first Processor
}

func (s *Splitter) SetFirst(first Processor) {
	s.first = first
}

func (s *Splitter) Process(ctx *FileContext) error {

	// Processors often change the FileContext, so we need to make a backup
	// and use that for the next processor. However the reader is a pointer,
	// and the state of the reader will definately change. However, the reader
	// implements io.ReadSeeker, so the next processor can still read from the
	// beginning. Processors must not close the reader.

	backup := &FileContext{
		Filename: ctx.Filename, // string
		Reader:   ctx.Reader,   // io.ReadSeeker
	}

	// Call the first processor
	slog.Info("Calling the first processor")
	err := s.first.Process(ctx)
	if err != nil {
		return fmt.Errorf("first processor failed: %s", err)
	}

	// Restore the reader and the filename
	ctx.Filename = backup.Filename
	ctx.Reader = backup.Reader

	// Seek to the beginning
	slog.Info("Seeking to the beginning")
	_, err = ctx.Reader.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to the beginning: %s", err)
	}

	// Call the second processor
	slog.Info("Calling the second processor")
	err = s.NextProcessor.Process(backup)
	if err != nil {
		return fmt.Errorf("second processor failed: %s", err)
	}

	slog.Info("Splitter done")
	return nil
}
