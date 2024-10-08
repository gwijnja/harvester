package harvester

import "io"

// NextProcessor is a struct that holds the next processor in the chain.
type NextProcessor struct {
	next FileWriter
}

// SetNext sets the next processor in the chain.
func (b *NextProcessor) SetNext(next FileWriter) {
	b.next = next
}

// Process calls the next processor in the chain, if it exists.
func (b *NextProcessor) Process(filename string, r io.Reader) error {
	if b.next != nil {
		return b.next.Process(filename, r)
	}
	return nil
}

// FileContext is a struct that contains a filename and a reader.
type FileContext struct {
	Filename string
	Reader   io.ReadSeeker
}
